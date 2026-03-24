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
			Use:   "rename <current> <new>",
			Short: "Rename a local remote alias, or set your first alias from a tray name/id",
			Long: `Changes remotes.json on your machine only.

If <current> matches an existing remote alias, it is renamed to <new> (same tray).

Otherwise <current> must be a tray name or id from tray ls (you are already a member); <new> becomes a new local alias for that tray.

This does not rename the tray on the server — only the owner can do that with tray rename.`,
			Args: cobra.ExactArgs(2),
			RunE: runRemoteRename,
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
