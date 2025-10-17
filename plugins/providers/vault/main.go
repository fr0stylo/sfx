package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	vault "github.com/hashicorp/vault/api"
	"gopkg.in/yaml.v3"

	"github.com/fr0stylo/sfx/provider"
)

type options struct {
	Address   string        `yaml:"address"`
	Token     string        `yaml:"token"`
	Namespace string        `yaml:"namespace"`
	Field     string        `yaml:"field"`
	Timeout   time.Duration `yaml:"timeout"`
}

func main() {
	provider.Run(provider.HandlerFunc(handle))
}

func handle(req provider.Request) (provider.Response, error) {
	var opts options
	if len(req.Options) > 0 {
		if err := yaml.Unmarshal(req.Options, &opts); err != nil {
			return provider.Response{}, fmt.Errorf("parse options: %w", err)
		}
	}

	addr := firstNonEmpty(opts.Address, os.Getenv("VAULT_ADDR"))
	token := firstNonEmpty(opts.Token, os.Getenv("VAULT_TOKEN"))

	if addr == "" {
		return provider.Response{}, errors.New("vault address not provided (set options.address or VAULT_ADDR)")
	}
	if token == "" {
		return provider.Response{}, errors.New("vault token not provided (set options.token or VAULT_TOKEN)")
	}

	config := vault.DefaultConfig()
	config.Address = addr
	if opts.Timeout > 0 {
		config.Timeout = opts.Timeout
	}

	client, err := vault.NewClient(config)
	if err != nil {
		return provider.Response{}, fmt.Errorf("create vault client: %w", err)
	}
	client.SetToken(token)
	if opts.Namespace != "" {
		client.SetNamespace(opts.Namespace)
	}

	path, field := splitRef(req.Ref)
	if path == "" {
		return provider.Response{}, errors.New("ref must include a vault path")
	}
	if field == "" {
		field = opts.Field
	}

	secret, err := client.Logical().Read(path)
	if err != nil {
		return provider.Response{}, fmt.Errorf("read secret %q: %w", path, err)
	}
	if secret == nil {
		return provider.Response{}, fmt.Errorf("secret %q not found", path)
	}

	value, err := extractValue(secret.Data, field)
	if err != nil {
		return provider.Response{}, fmt.Errorf("extract value: %w", err)
	}

	return provider.Response{Value: value}, nil
}

func splitRef(ref string) (string, string) {
	parts := strings.SplitN(ref, "#", 2)
	path := strings.TrimSpace(parts[0])
	if len(parts) == 2 {
		return path, strings.TrimSpace(parts[1])
	}
	return path, ""
}

func extractValue(data map[string]any, field string) ([]byte, error) {
	if nested, ok := data["data"].(map[string]any); ok {
		data = nested
	}

	if field == "" {
		if len(data) != 1 {
			return nil, errors.New("field must be specified (ref '#field' or options.field)")
		}
		for _, v := range data {
			return formatValue(v), nil
		}
		return nil, errors.New("secret data empty")
	}

	val, ok := data[field]
	if !ok {
		return nil, fmt.Errorf("field %q not found", field)
	}
	return formatValue(val), nil
}

func formatValue(v any) []byte {
	switch t := v.(type) {
	case nil:
		return nil
	case []byte:
		return t
	case string:
		return []byte(t)
	default:
		return []byte(fmt.Sprint(t))
	}
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
