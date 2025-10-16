package main

import (
	"fmt"

	"sfx/plugin"
)

func main() {
	plugin.Run(plugin.HandlerFunc(handle))
}

func handle(req plugin.Request) (plugin.Response, error) {
	value := fmt.Sprintf("secret-for-%s", req.Ref)
	return plugin.Response{Value: []byte(value)}, nil
}
