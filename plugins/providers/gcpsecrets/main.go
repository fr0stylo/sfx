package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"gopkg.in/yaml.v3"

	"github.com/fr0stylo/sfx/provider"
)

const defaultGCPSecretTimeout = 30 * time.Second

type options struct {
	Project string        `yaml:"project"`
	Secret  string        `yaml:"secret"`
	Version string        `yaml:"version"`
	Timeout time.Duration `yaml:"timeout"`
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

	name, err := resolveResource(req.Ref, opts)
	if err != nil {
		return provider.Response{}, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), resolveTimeout(opts.Timeout, defaultGCPSecretTimeout))
	defer cancel()

	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return provider.Response{}, fmt.Errorf("create secret manager client: %w", err)
	}
	defer client.Close()

	resp, err := client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
		Name: name,
	})
	if err != nil {
		return provider.Response{}, fmt.Errorf("access %q: %w", name, err)
	}

	return provider.Response{Value: resp.GetPayload().GetData()}, nil
}

func resolveResource(ref string, opts options) (string, error) {
	ref = strings.TrimSpace(ref)
	if strings.HasPrefix(ref, "projects/") {
		if strings.Contains(ref, "/versions/") {
			return ref, nil
		}
		version := opts.Version
		if version == "" {
			version = "latest"
		}
		return fmt.Sprintf("%s/versions/%s", strings.TrimSuffix(ref, "/"), version), nil
	}

	secretPart, versionPart := splitRef(ref)
	if secretPart == "" {
		secretPart = opts.Secret
	}
	version := firstNonEmpty(versionPart, opts.Version, "latest")

	project := opts.Project
	if secretPart != "" {
		if strings.Contains(secretPart, "/") {
			parts := strings.SplitN(secretPart, "/", 2)
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				return "", fmt.Errorf("invalid secret ref %q", secretPart)
			}
			project = parts[0]
			secretPart = parts[1]
		}
	}

	if secretPart == "" {
		return "", errors.New("secret name missing (ref or options.secret)")
	}
	if project == "" {
		return "", errors.New("project missing (options.project or ref 'project/secret')")
	}

	return fmt.Sprintf("projects/%s/secrets/%s/versions/%s", project, secretPart, version), nil
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
