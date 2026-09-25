package commands

import (
	"time"

	"github.com/spf13/cobra"
)

func cmdBar() *cobra.Command {
	c := &cobra.Command{
		Use:   "bar",
		Short: "Show a macOS menu bar icon for open/pending tray items",
		RunE:  runBar,
	}
	c.Flags().Duration("interval", 30*time.Second, "refresh interval")
	c.Flags().Bool("daemon", false, "daemon mode: write bar.pid and bar.log under config dir")
	return c
}

func runBar(cmd *cobra.Command, args []string) error {
	return runBarUI(cmd)
}
