package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"gopkg.in/yaml.v3"

	"github.com/fr0stylo/sfx/provider"
)

const defaultAWSSecretsTimeout = 30 * time.Second

type options struct {
	Region       string        `yaml:"region"`
	Profile      string        `yaml:"profile"`
	VersionID    string        `yaml:"version_id"`
	VersionStage string        `yaml:"version_stage"`
	Timeout      time.Duration `yaml:"timeout"`
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

	secretID, versionID, versionStage := parseRef(req.Ref)
	if secretID == "" {
		return provider.Response{}, errors.New("ref must include the secret identifier")
	}

	if opts.VersionID == "" && versionID != "" {
		opts.VersionID = versionID
	}
	if opts.VersionStage == "" && versionStage != "" {
		opts.VersionStage = versionStage
	}

	ctx, cancel := context.WithTimeout(context.Background(), resolveTimeout(opts.Timeout, defaultAWSSecretsTimeout))
	defer cancel()

	cfgOpts := []func(*config.LoadOptions) error{}
	if opts.Region != "" {
		cfgOpts = append(cfgOpts, config.WithRegion(opts.Region))
	}
	if opts.Profile != "" {
		cfgOpts = append(cfgOpts, config.WithSharedConfigProfile(opts.Profile))
	}

	cfg, err := config.LoadDefaultConfig(ctx, cfgOpts...)
	if err != nil {
		return provider.Response{}, fmt.Errorf("load aws config: %w", err)
	}

	client := secretsmanager.NewFromConfig(cfg)
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretID),
	}
	if opts.VersionID != "" {
		input.VersionId = aws.String(opts.VersionID)
	}
	if opts.VersionStage != "" {
		input.VersionStage = aws.String(opts.VersionStage)
	}

	resp, err := client.GetSecretValue(ctx, input)
	if err != nil {
		return provider.Response{}, fmt.Errorf("get secret %q: %w", secretID, err)
	}

	switch {
	case resp.SecretString != nil:
		return provider.Response{Value: []byte(*resp.SecretString)}, nil
	case len(resp.SecretBinary) > 0:
		return provider.Response{Value: resp.SecretBinary}, nil
	default:
		return provider.Response{}, errors.New("secret contained no data")
	}
}

func parseRef(ref string) (secretID, versionID, versionStage string) {
	part := strings.TrimSpace(ref)
	if part == "" {
		return "", "", ""
	}

	segments := strings.SplitN(part, "#", 2)
	secretID = strings.TrimSpace(segments[0])

	if len(segments) == 2 {
		meta := strings.TrimSpace(segments[1])
		switch {
		case strings.HasPrefix(meta, "version:"):
			versionID = strings.TrimSpace(strings.TrimPrefix(meta, "version:"))
		case strings.HasPrefix(meta, "version_id:"):
			versionID = strings.TrimSpace(strings.TrimPrefix(meta, "version_id:"))
		case strings.HasPrefix(meta, "stage:"):
			versionStage = strings.TrimSpace(strings.TrimPrefix(meta, "stage:"))
		case strings.HasPrefix(meta, "version_stage:"):
			versionStage = strings.TrimSpace(strings.TrimPrefix(meta, "version_stage:"))
		default:
			// Default to treating the suffix as a stage for convenience.
			versionStage = meta
		}
	}

	return secretID, versionID, versionStage
}

func resolveTimeout(given time.Duration, fallback time.Duration) time.Duration {
	if given <= 0 {
		return fallback
	}
	return given
}
