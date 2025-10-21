package config

import (
	"errors"
	"strings"
	"testing"
)

func TestValidateSuccess(t *testing.T) {
	cfg := Config{
		Providers: map[string]string{
			"vault": "./bin/providers/vault",
		},
		Exporters: map[string]string{
			"env": "./bin/exporters/env",
		},
		Output: Output{
			Type: "env",
		},
		Secrets: map[string]Secret{
			"DB_PASSWORD": {
				Ref:      "secret/data/app#password",
				Provider: "vault",
			},
		},
	}

	if err := Validate(cfg); err != nil {
		t.Fatalf("Validate returned error for valid config: %v", err)
	}
}

func TestValidateReportsIssues(t *testing.T) {
	cfg := Config{
		Providers: map[string]string{
			"vault": "",
		},
		Exporters: map[string]string{},
		Output: Output{
			Type: "shell",
		},
		Secrets: map[string]Secret{
			"": {
				Ref:      "",
				Provider: "aws",
			},
		},
	}

	err := Validate(cfg)
	if err == nil {
		t.Fatalf("Validate returned nil, want error")
	}

	var vErr ValidationError
	if !errors.As(err, &vErr) {
		t.Fatalf("expected ValidationError, got %T", err)
	}

	if len(vErr.Issues) == 0 {
		t.Fatalf("expected issues, got 0")
	}

	wantSubstrings := []string{
		"provider \"vault\" has an empty binary path",
		"no exporters configured",
		"output.type \"shell\" does not match any configured exporter",
		"secret name cannot be empty",
		"secret \"\" is missing ref",
		"secret \"\" references unknown provider \"aws\"",
	}

	for _, want := range wantSubstrings {
		found := false
		for _, issue := range vErr.Issues {
			if strings.Contains(issue, want) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected an issue containing %q, issues: %v", want, vErr.Issues)
		}
	}
}
