package cli

import (
	"github.com/spf13/cobra"
)

func cmdRotateInvite() *cobra.Command {
	return &cobra.Command{
		Use:   "rotate-invite <tray-name>",
		Short: "Issue a new invite token (same as `tray invite <tray> --rotate`)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInviteCore(cmd, args, true)
		},
	}
}
