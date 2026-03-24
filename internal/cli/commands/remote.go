package commands

import "github.com/spf13/cobra"

func cmdRemote() *cobra.Command {
	root := &cobra.Command{
		Use:   "remote",
		Short: "Manage local aliases for trays (join + remember a short name)",
	}

	root.AddCommand(
		&cobra.Command{
			Use:   "add <alias> <invite-url-or-token>",
			Short: "Join a tray via invite and save a local alias",
			Long:  `Runs the same join as tray join, then stores alias → tray id in remotes.json under your tray config directory.`,
			Args:  cobra.ExactArgs(2),
			RunE:  runRemoteAdd,
		},
		&cobra.Command{
			Use:   "link <alias> <tray>",
			Short: "Save a local alias for a tray you already have (no invite needed)",
			Long: `Stores alias → tray id in remotes.json. Use this when you are already a member (e.g. after tray join) and only want a short local name.

This does not rename the tray on the server — only the owner can do that with tray rename.`,
			Args: cobra.ExactArgs(2),
			RunE: runRemoteLink,
		},
		&cobra.Command{
			Use:   "ls",
			Short: "List remotes",
			RunE:  runRemoteLs,
		},
		&cobra.Command{
			Use:   "remove <alias>",
			Short: "Remove a remote alias",
			Args:  cobra.ExactArgs(1),
			RunE:  runRemoteRemove,
		},
	)

	return root
}
