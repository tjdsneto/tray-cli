package main

import (
	"fmt"
	"os"

	"github.com/tjdsneto/tray-cli/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "tray: %v\n", err)
		os.Exit(1)
	}
}
