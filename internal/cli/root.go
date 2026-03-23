package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tjdsneto/tray-cli/internal/output"
)

// Execute runs the tray root command.
func Execute() error {
	return NewRootCmd().ExecuteContext(context.Background())
}

// NewRootCmd builds the full CLI tree (stubs until each feature lands).
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "tray",
		Short:         "Tray-CLI — shared inbox tray (Supabase backend)",
		SilenceUsage:  true,
		SilenceErrors: true,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	root.PersistentFlags().StringVar(&configDirFlag, "config-dir", "", "override config directory (default: $XDG_CONFIG_HOME/tray or ~/.config/tray)")
	output.RegisterPersistentFlags(root.PersistentFlags())

	root.AddCommand(
		cmdLogin(),
		cmdStatus(),
		cmdCreate(),
		cmdLs(),
		cmdRename(),
		cmdDeleteTray(),
		cmdInvite(),
		cmdRotateInvite(),
		cmdJoin(),
		cmdAdd(),
		cmdList(),
		cmdRemove(),
		cmdContributed(),
		cmdRemote(),
		cmdMembers(),
		cmdRevoke(),
		cmdLeave(),
		cmdNotImplemented("review", "Interactive triage"),
		cmdAccept(),
		cmdDecline(),
		cmdSnooze(),
		cmdComplete(),
		cmdArchive(),
		cmdNotImplemented("listen", "Watch for tray updates"),
	)

	return root
}

var configDirFlag string

func cmdNotImplemented(use, short string) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("`tray %s` isn't available yet — run `tray help` for supported commands", use)
		},
	}
}
