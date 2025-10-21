package config

import (
	"fmt"
	"sort"
	"strings"
)

// ValidationError contains the list of configuration issues discovered during validation.
type ValidationError struct {
	Issues []string
}

// Error implements the error interface.
func (e ValidationError) Error() string {
	switch len(e.Issues) {
	case 0:
		return "configuration validation failed"
	case 1:
		return e.Issues[0]
	default:
		return fmt.Sprintf("%d configuration errors:\n- %s", len(e.Issues), strings.Join(e.Issues, "\n- "))
	}
}

// Validate performs a series of consistency checks on the loaded configuration.
func Validate(cfg Config) error {
	var issues []string

	if len(cfg.Providers) == 0 {
		issues = append(issues, "no providers configured")
	} else {
		providerNames := sortedKeys(cfg.Providers)
		for _, name := range providerNames {
			if strings.TrimSpace(cfg.Providers[name]) == "" {
				issues = append(issues, fmt.Sprintf("provider %q has an empty binary path", name))
			}
		}
	}

	if len(cfg.Exporters) == 0 {
		issues = append(issues, "no exporters configured")
	} else {
		exporterNames := sortedKeys(cfg.Exporters)
		for _, name := range exporterNames {
			if strings.TrimSpace(cfg.Exporters[name]) == "" {
				issues = append(issues, fmt.Sprintf("exporter %q has an empty binary path", name))
			}
		}
	}

	outputType := strings.TrimSpace(cfg.Output.Type)
	if outputType == "" {
		issues = append(issues, "output.type must be specified")
	} else if _, ok := cfg.Exporters[outputType]; !ok {
		issues = append(issues, fmt.Sprintf("output.type %q does not match any configured exporter", outputType))
	}

	if len(cfg.Secrets) > 0 {
		secretNames := sortedKeys(cfg.Secrets)
		for _, name := range secretNames {
			secret := cfg.Secrets[name]
			if strings.TrimSpace(name) == "" {
				issues = append(issues, "secret name cannot be empty")
			}
			if strings.TrimSpace(secret.Ref) == "" {
				issues = append(issues, fmt.Sprintf("secret %q is missing ref", name))
			}
			providerName := strings.TrimSpace(secret.Provider)
			if providerName == "" {
				issues = append(issues, fmt.Sprintf("secret %q is missing provider", name))
			} else if _, ok := cfg.Providers[providerName]; !ok {
				issues = append(issues, fmt.Sprintf("secret %q references unknown provider %q", name, providerName))
			}
		}
	}

	if len(issues) > 0 {
		return ValidationError{Issues: issues}
	}
	return nil
}

func sortedKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
