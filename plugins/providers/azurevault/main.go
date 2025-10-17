package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"gopkg.in/yaml.v3"

	"sfx/plugin"
)

const defaultAzureVaultTimeout = 30 * time.Second

type options struct {
	VaultURL  string        `yaml:"vault_url"`
	VaultName string        `yaml:"vault_name"`
	Secret    string        `yaml:"secret"`
	Version   string        `yaml:"version"`
	Timeout   time.Duration `yaml:"timeout"`
}

func main() {
	plugin.Run(plugin.HandlerFunc(handle))
}

func handle(req plugin.Request) (plugin.Response, error) {
	var opts options
	if len(req.Options) > 0 {
		if err := yaml.Unmarshal(req.Options, &opts); err != nil {
			return plugin.Response{}, fmt.Errorf("parse options: %w", err)
		}
	}

	vaultURL, secretName, version, err := resolveTarget(req.Ref, opts)
	if err != nil {
		return plugin.Response{}, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), resolveTimeout(opts.Timeout, defaultAzureVaultTimeout))
	defer cancel()

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return plugin.Response{}, fmt.Errorf("obtain azure credential: %w", err)
	}

	client, err := azsecrets.NewClient(vaultURL, cred, nil)
	if err != nil {
		return plugin.Response{}, fmt.Errorf("create key vault client: %w", err)
	}

	resp, err := client.GetSecret(ctx, secretName, version, nil)
	if err != nil {
		return plugin.Response{}, fmt.Errorf("get secret %q: %w", secretName, err)
	}

	if resp.Value == nil {
		return plugin.Response{}, errors.New("secret value empty")
	}

	return plugin.Response{Value: []byte(*resp.Value)}, nil
}

func resolveTarget(ref string, opts options) (string, string, string, error) {
	path, version := splitRef(strings.TrimSpace(ref))

	if path != "" && strings.HasPrefix(path, "http") {
		u, err := url.Parse(path)
		if err != nil {
			return "", "", "", fmt.Errorf("parse vault url: %w", err)
		}
		segments := strings.Split(strings.Trim(u.Path, "/"), "/")
		if len(segments) < 2 || !strings.EqualFold(segments[0], "secrets") {
			return "", "", "", fmt.Errorf("invalid azure vault path %q", path)
		}
		secretName := segments[1]
		if version == "" && len(segments) > 2 {
			version = segments[2]
		}
		return ensureVaultURL(u.Scheme + "://" + u.Host), secretName, firstNonEmpty(version, opts.Version, ""), nil
	}

	vaultURL := strings.TrimSpace(opts.VaultURL)
	secretName := strings.TrimSpace(opts.Secret)

	if path != "" {
		trimmed := strings.Trim(path, "/")
		if trimmed != "" {
			parts := strings.Split(trimmed, "/")
			switch len(parts) {
			case 1:
				secretName = parts[0]
			default:
				if secretName == "" {
					secretName = parts[1]
				}
				if vaultURL == "" {
					vaultURL = inferVaultURL(parts[0], opts.VaultName)
				}
				if len(parts) > 2 && version == "" {
					version = parts[2]
				}
			}
		}
	}

	if vaultURL == "" && opts.VaultName != "" {
		vaultURL = inferVaultURL(opts.VaultName, "")
	}
	if secretName == "" {
		return "", "", "", errors.New("secret name missing (ref or options.secret)")
	}
	version = firstNonEmpty(version, opts.Version)

	if vaultURL == "" {
		return "", "", "", errors.New("vault url not provided (options.vault_url, options.vault_name, or ref 'vault/secret')")
	}

	return ensureVaultURL(vaultURL), secretName, version, nil
}

func ensureVaultURL(raw string) string {
	u := strings.TrimRight(raw, "/")
	return u
}

func inferVaultURL(vaultRef string, fallback string) string {
	name := strings.TrimSpace(vaultRef)
	if name == "" {
		name = strings.TrimSpace(fallback)
	}
	if name == "" {
		return ""
	}
	if strings.HasPrefix(name, "http") {
		return ensureVaultURL(name)
	}
	name = strings.TrimSuffix(name, ".vault.azure.net")
	return fmt.Sprintf("https://%s.vault.azure.net", name)
}

func splitRef(ref string) (string, string) {
	if ref == "" {
		return "", ""
	}
	parts := strings.SplitN(ref, "#", 2)
	head := strings.TrimSpace(parts[0])
	if len(parts) == 2 {
		return head, strings.TrimSpace(parts[1])
	}
	return head, ""
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func resolveTimeout(given time.Duration, fallback time.Duration) time.Duration {
	if given <= 0 {
		return fallback
	}
	return given
}
