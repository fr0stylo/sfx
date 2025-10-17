package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"sfx/config"
	"sfx/internal/client"
	"sfx/internal/rpc"
)

func init() {
	RegisterSubCommand(newFetchCommand())
}

func newFetchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fetch",
		Short: "Fetch secrets and render output",
		Long:  "Fetch secrets from configured providers and render them using the configured exporter.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFetch(cmd.Context(), os.Stdout)
		},
	}

	cmd.Flags().StringP("output", "o", "env", "Output type (env, tfvars, template, shell, k8ssecret, ansible)")
	// TODO: Find out how to do this properly
	//cmd.Flags().String("output-option", "", "Override output options (key=value or key=value;key=value)")
	cmd.Flags().String("output-template", "", "Output options (key=value or key=value;key=value)")

	Must(viper.BindPFlag("output.type", cmd.Flags().Lookup("output")))
	// TODO: Find out how to do this properly
	//Must(viper.BindPFlag("output.options", cmd.Flags().Lookup("output-option")))
	Must(viper.BindPFlag("output.template", cmd.Flags().Lookup("output-template")))

	return cmd
}

func runFetch(ctx context.Context, out io.Writer) error {
	slog.SetLogLoggerLevel(slog.LevelInfo)

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	secrets := map[string][]byte{}
	for name, secret := range cfg.Secrets {
		providerPath, ok := cfg.Providers[secret.Provider]
		if !ok || providerPath == "" {
			return fmt.Errorf("provider %q not configured", secret.Provider)
		}

		val, err := fetchSecret(ctx, providerPath, secret.Ref, secret.ProviderOptions)
		if err != nil {
			return fmt.Errorf("fetch %q: %w", name, err)
		}

		if _, exists := secrets[name]; exists {
			slog.Warn("duplicate secret name", "name", name)
		}

		secrets[name] = val
	}

	targetOutput := cfg.Output.Type
	exporterPath := cfg.Exporters[targetOutput]
	if exporterPath == "" {
		return fmt.Errorf("exporter for type %q not configured", targetOutput)
	}

	options := cfg.Output.Options
	rawOption := strings.TrimSpace(viper.GetString("fetch.option"))
	if rawOption != "" {
		parsed, err := parseOptionOverride(rawOption)
		if err != nil {
			return err
		}
		for k, v := range parsed {
			if options == nil {
				options = map[string]any{}
			}
			options[k] = v
		}
	}

	data, err := formatSecrets(ctx, exporterPath, secrets, options)
	if err != nil {
		return fmt.Errorf("format output: %w", err)
	}

	if _, err := io.Copy(out, bytes.NewReader(data)); err != nil {
		return fmt.Errorf("write output: %w", err)
	}

	return nil
}

func fetchSecret(ctx context.Context, path, ref string, options map[string]any) ([]byte, error) {
	opts, err := marshalOptions(options)
	if err != nil {
		return nil, err
	}

	req := &rpc.SecretRequest{Ref: ref, Options: opts}
	var resp rpc.SecretResponse
	if err := client.Call(ctx, path, req, &resp); err != nil {
		return nil, err
	}

	if resp.Error != "" {
		return nil, fmt.Errorf("provider error: %s", resp.Error)
	}
	return resp.Value, nil
}

func formatSecrets(ctx context.Context, path string, data map[string][]byte, options map[string]any) ([]byte, error) {
	opts, err := marshalOptions(options)
	if err != nil {
		return nil, err
	}

	req := &rpc.ExportRequest{Values: data, Options: opts}
	var resp rpc.ExportResponse
	if err := client.Call(ctx, path, req, &resp); err != nil {
		return nil, err
	}

	if resp.Error != "" {
		return nil, fmt.Errorf("exporter error: %s", resp.Error)
	}

	return resp.GetPayload(), nil
}

func marshalOptions(options map[string]any) ([]byte, error) {
	if len(options) == 0 {
		return nil, nil
	}
	return yaml.Marshal(options)
}

func parseOptionOverride(raw string) (map[string]any, error) {
	result := make(map[string]any)
	for _, item := range strings.Split(raw, ";") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		parts := strings.SplitN(item, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid option %q (expected key=value)", item)
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" {
			return nil, fmt.Errorf("invalid option %q (empty key)", item)
		}
		result[key] = value
	}
	return result, nil
}
