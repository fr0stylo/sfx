package main

import (
	"fmt"

	"gopkg.in/yaml.v3"

	"sfx/plugin"
)

type Options struct {
	Path string `yaml:"path"`
}

func main() {
	plugin.Run(plugin.HandlerFunc(handle))
}

func handle(req plugin.Request) (plugin.Response, error) {
	var opts Options
	if err := yaml.Unmarshal(req.Options, &opts); err != nil {
		return plugin.Response{}, err
	}

	value := fmt.Sprintf("secret-for-%s", req.Ref)
	return plugin.Response{Value: []byte(opts.Path + value)}, nil
}
