package main

import (
	"encoding/base64"
	"fmt"
	"sort"

	"gopkg.in/yaml.v3"

	"github.com/fr0stylo/sfx/exporter"
)

type options struct {
	Name        string            `yaml:"name"`
	Namespace   string            `yaml:"namespace"`
	Type        string            `yaml:"type"`
	Labels      map[string]string `yaml:"labels"`
	Annotations map[string]string `yaml:"annotations"`
}

type secretManifest struct {
	APIVersion string            `yaml:"apiVersion"`
	Kind       string            `yaml:"kind"`
	Metadata   metadata          `yaml:"metadata"`
	Type       string            `yaml:"type,omitempty"`
	Data       map[string]string `yaml:"data"`
}

type metadata struct {
	Name        string            `yaml:"name"`
	Namespace   string            `yaml:"namespace,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

func main() {
	exporter.Run(exporter.HandlerFunc(handle))
}

func handle(req exporter.Request) (exporter.Response, error) {
	var opts options
	if len(req.Options) > 0 {
		if err := yaml.Unmarshal(req.Options, &opts); err != nil {
			return exporter.Response{}, fmt.Errorf("parse options: %w", err)
		}
	}

	if opts.Name == "" {
		return exporter.Response{}, fmt.Errorf("option name is required")
	}

	data := make(map[string]string, len(req.Values))
	for key, value := range req.Values {
		data[key] = base64.StdEncoding.EncodeToString(value)
	}

	manifest := secretManifest{
		APIVersion: "v1",
		Kind:       "Secret",
		Metadata: metadata{
			Name:        opts.Name,
			Namespace:   opts.Namespace,
			Labels:      opts.Labels,
			Annotations: opts.Annotations,
		},
		Type: opts.Type,
		Data: orderedMap(data),
	}

	payload, err := yaml.Marshal(manifest)
	if err != nil {
		return exporter.Response{}, fmt.Errorf("marshal secret yaml: %w", err)
	}

	return exporter.Response{Payload: payload}, nil
}

func orderedMap(m map[string]string) map[string]string {
	if len(m) == 0 {
		return nil
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	out := make(map[string]string, len(m))
	for _, k := range keys {
		out[k] = m[k]
	}
	return out
}
