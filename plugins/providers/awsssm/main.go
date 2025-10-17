package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"gopkg.in/yaml.v3"

	"github.com/fr0stylo/sfx/provider"
)

const defaultAWSSSMTimeout = 30 * time.Second

type options struct {
	Region         string        `yaml:"region"`
	Profile        string        `yaml:"profile"`
	WithDecryption *bool         `yaml:"with_decryption"`
	Timeout        time.Duration `yaml:"timeout"`
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

	paramName := strings.TrimSpace(req.Ref)
	if paramName == "" {
		return provider.Response{}, errors.New("ref must include the parameter name")
	}

	ctx, cancel := context.WithTimeout(context.Background(), resolveTimeout(opts.Timeout, defaultAWSSSMTimeout))
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

	client := ssm.NewFromConfig(cfg)

	var withDecryption = true
	if opts.WithDecryption != nil {
		withDecryption = *opts.WithDecryption
	}

	resp, err := client.GetParameter(ctx, &ssm.GetParameterInput{
		Name:           &paramName,
		WithDecryption: &withDecryption,
	})
	if err != nil {
		return provider.Response{}, fmt.Errorf("get parameter %q: %w", paramName, err)
	}
	if resp.Parameter == nil {
		return provider.Response{}, errors.New("parameter missing in response")
	}
	if resp.Parameter.Value == nil {
		return provider.Response{}, errors.New("parameter value empty")
	}

	return provider.Response{Value: []byte(*resp.Parameter.Value)}, nil
}

func resolveTimeout(given time.Duration, fallback time.Duration) time.Duration {
	if given <= 0 {
		return fallback
	}
	return given
}
