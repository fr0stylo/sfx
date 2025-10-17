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

	"sfx/plugin"
)

const defaultAWSSSMTimeout = 30 * time.Second

type options struct {
	Region         string        `yaml:"region"`
	Profile        string        `yaml:"profile"`
	WithDecryption *bool         `yaml:"with_decryption"`
	Timeout        time.Duration `yaml:"timeout"`
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

	paramName := strings.TrimSpace(req.Ref)
	if paramName == "" {
		return plugin.Response{}, errors.New("ref must include the parameter name")
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
		return plugin.Response{}, fmt.Errorf("load aws config: %w", err)
	}

	client := ssm.NewFromConfig(cfg)

	var withDecryption bool = true
	if opts.WithDecryption != nil {
		withDecryption = *opts.WithDecryption
	}

	resp, err := client.GetParameter(ctx, &ssm.GetParameterInput{
		Name:           &paramName,
		WithDecryption: &withDecryption,
	})
	if err != nil {
		return plugin.Response{}, fmt.Errorf("get parameter %q: %w", paramName, err)
	}
	if resp.Parameter == nil {
		return plugin.Response{}, errors.New("parameter missing in response")
	}
	if resp.Parameter.Value == nil {
		return plugin.Response{}, errors.New("parameter value empty")
	}

	return plugin.Response{Value: []byte(*resp.Parameter.Value)}, nil
}

func resolveTimeout(given time.Duration, fallback time.Duration) time.Duration {
	if given <= 0 {
		return fallback
	}
	return given
}
