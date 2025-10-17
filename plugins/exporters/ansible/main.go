package main

import (
	"fmt"
	"sort"

	"gopkg.in/yaml.v3"

	"sfx/exporter"
)

type options struct {
	Prefix string   `yaml:"prefix"`
	Order  []string `yaml:"order"`
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

	keys := orderedKeys(req.Values, opts.Order)

	values := make(map[string]string, len(req.Values))
	for _, key := range keys {
		name := key
		if opts.Prefix != "" {
			name = opts.Prefix + key
		}
		values[name] = string(req.Values[key])
	}

	payload, err := yaml.Marshal(values)
	if err != nil {
		return exporter.Response{}, fmt.Errorf("marshal yaml: %w", err)
	}

	return exporter.Response{Payload: payload}, nil
}

func orderedKeys(values map[string][]byte, order []string) []string {
	seen := make(map[string]bool, len(values))
	var out []string

	for _, key := range order {
		if _, ok := values[key]; ok && !seen[key] {
			out = append(out, key)
			seen[key] = true
		}
	}

	var remaining []string
	for key := range values {
		if !seen[key] {
			remaining = append(remaining, key)
		}
	}
	sort.Strings(remaining)
	out = append(out, remaining...)
	return out
}
