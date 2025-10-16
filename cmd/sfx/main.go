package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"

	"sfx/config"
	"sfx/internal/client"
	"sfx/internal/rpc"
)

func main() {
	buf, err := os.ReadFile(".sfx.yaml")
	if err != nil && errors.Is(err, os.ErrNotExist) {
		fmt.Println("no sfx.yaml found in current directory")
		os.Exit(1)
	}

	var cfg config.Config
	if err := yaml.Unmarshal(buf, &cfg); err != nil {
		fmt.Println("error parsing sfx.yaml:", err)
		os.Exit(1)
	}

	secrets := map[string][]byte{}
	for n, s := range cfg.Secrets {
		fmt.Println(s.Ref, s.Provider)

		providerPath, ok := cfg.Providers[s.Provider]
		if !ok {
			fmt.Println("provider not found ", s.Provider, " in providers ", cfg.Providers)
			os.Exit(1)
		}
		val, err := fetch(providerPath, s.Ref, s.ProviderOptions)
		if err != nil {
			fmt.Println("error fetching secret:", err)
			os.Exit(1)
		}

		if secrets[n] != nil {
			fmt.Println("duplicate secret name:", n)
		}

		secrets[n] = val
	}

	data, err := format("./bin/env", secrets, cfg.Output.OutputOptions)
	if err != nil {
		fmt.Println("error formatting output:", err)
		os.Exit(1)
	}

	_, _ = io.Copy(os.Stdout, bytes.NewReader(data))
}

func fetch(path, ref string, options any) ([]byte, error) {
	opts, err := yaml.Marshal(options)
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

	return resp.Value, nil
}

func format(path string, data map[string][]byte, options []byte) ([]byte, error) {
	req := &rpc.ExportRequest{Values: data, Options: options}
	var resp rpc.ExportResponse
	if err := client.Call(path, req, &resp); err != nil {
		return nil, err
	}

	if resp.Error != "" {
		return nil, fmt.Errorf("exporter error: %s", resp.Error)
	}

	return resp.GetPayload(), nil
}
