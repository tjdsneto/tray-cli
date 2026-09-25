//go:build !darwin

package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func runBarUI(cmd *cobra.Command) error {
	return fmt.Errorf("tray bar is only supported on macOS")
}
