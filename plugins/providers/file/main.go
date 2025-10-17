package main

import (
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/fr0stylo/sfx/provider"
)

type Options struct {
	Path string `yaml:"path"`
}

func main() {
	provider.Run(provider.HandlerFunc(handle))
}

func handle(req provider.Request) (provider.Response, error) {
	var opts Options
	if err := yaml.Unmarshal(req.Options, &opts); err != nil {
		return provider.Response{}, err
	}

	value := fmt.Sprintf("secret-for-%s", req.Ref)
	return provider.Response{Value: []byte(opts.Path + value)}, nil
}
