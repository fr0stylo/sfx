package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"

	"gopkg.in/yaml.v3"

	"sfx/config"
	"sfx/internal/client"
	"sfx/internal/rpc"
)

var (
	logger = log.New(os.Stdout, "ERROR: ", log.LstdFlags)
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelInfo)
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("failed to load configuration: ", err)
	}

	secrets := map[string][]byte{}
	slog.Debug("Secrets", "secrets", cfg.Secrets)
	for name, secret := range cfg.Secrets {
		providerPath, ok := cfg.Providers[secret.Provider]
		if !ok || providerPath == "" {
			logger.Fatal("provider not found ", secret.Provider, " in providers ", cfg.Providers)
		}

		val, err := fetch(providerPath, secret.Ref, secret.ProviderOptions)
		if err != nil {
			logger.Fatal("error fetching secret: ", err)
		}

		if secrets[name] != nil {
			logger.Println("duplicate secret name: ", name)
		}

		secrets[name] = val
	}

	exporterPath := cfg.Exporters[cfg.Output.Type]
	if exporterPath == "" {
		logger.Fatal("exporter not found for type: ", cfg.Output.Type)
	}

	data, err := format(exporterPath, secrets, cfg.Output.Options)
	if err != nil {
		logger.Fatal("error formatting output: ", err)
	}

	_, _ = io.Copy(os.Stdout, bytes.NewReader(data))
}

func fetch(path, ref string, options map[string]any) ([]byte, error) {
	slog.Debug("fetching secret", "ref", ref, "path", path)
	opts, err := marshalOptions(options)
	if err != nil {
		return nil, err
	}

	req := &rpc.SecretRequest{Ref: ref, Options: opts}
	var resp rpc.SecretResponse
	if err := client.Call(path, req, &resp); err != nil {
		return nil, err
	}

	if resp.Error != "" {
		return nil, fmt.Errorf("provider error: %s", resp.Error)
	}
	slog.Debug("fetching result", "ref", ref, "path", path, "value", string(resp.Value))

	return resp.Value, nil
}

func format(path string, data map[string][]byte, options map[string]any) ([]byte, error) {
	slog.Debug("formatting", "path", path, "data", data, "options", options)

	opts, err := marshalOptions(options)
	if err != nil {
		return nil, err
	}

	req := &rpc.ExportRequest{Values: data, Options: opts}
	var resp rpc.ExportResponse
	if err := client.Call(path, req, &resp); err != nil {
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
