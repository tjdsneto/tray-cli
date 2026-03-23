package main

import (
	"fmt"
	"os"

	"github.com/tjdsneto/tray-cli/internal/cli"
	"github.com/tjdsneto/tray-cli/internal/config"
)

func main() {
	if err := cli.Execute(); err != nil {
		if config.Debug() {
			fmt.Fprintf(os.Stderr, "tray [debug] %v\n", err)
		}
		fmt.Fprintf(os.Stderr, "tray: %s\n", cli.UserFacingError(err))
		os.Exit(1)
	}
}
