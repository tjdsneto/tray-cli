package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/tjdsneto/tray-cli/internal/cli"
	"github.com/tjdsneto/tray-cli/internal/config"
)

func init() {
	// Optional .env files (no error if missing). Cwd first, then config dir.
	_ = godotenv.Load()
	_ = godotenv.Load(filepath.Join(config.DefaultConfigDir(), ".env"))
}

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "tray: %v\n", err)
		os.Exit(1)
	}
}
