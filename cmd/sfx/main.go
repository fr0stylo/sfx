package main

import (
	"context"

	"sfx/cmd/sfx/cmd"
)

func main() {
	cmd.Must(cmd.Execute(context.Background()))
}
