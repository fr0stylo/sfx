package main

import (
	"context"

	"github.com/fr0stylo/sfx/cmd/cmd"
)

func main() {
	cmd.Must(cmd.Execute(context.Background()))
}
