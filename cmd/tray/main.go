package main

import (
	"os"

	"github.com/tjdsneto/tray-cli/internal/cli"
	"github.com/tjdsneto/tray-cli/internal/config"
)

func main() {
	if err := cli.Execute(); err != nil {
		cli.WriteUserError(os.Stderr, err, config.Debug())
		os.Exit(1)
	}
}
