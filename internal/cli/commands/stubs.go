package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func cmdNotImplemented(use, short string) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("`tray %s` isn't available yet — run `tray help` for supported commands", use)
		},
	}
}
