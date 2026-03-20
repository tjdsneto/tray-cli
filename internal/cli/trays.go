package cli

import "github.com/spf13/cobra"

func cmdCreate() *cobra.Command {
	return &cobra.Command{
		Use:   "create <name>",
		Short: "Create a named tray",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return stub("create")
		},
	}
}

func cmdLs() *cobra.Command {
	return &cobra.Command{
		Use:     "ls",
		Aliases: []string{"list-trays"},
		Short:   "List trays you own",
		RunE: func(cmd *cobra.Command, args []string) error {
			return stub("ls")
		},
	}
}

func cmdInvite() *cobra.Command {
	return &cobra.Command{
		Use:   "invite <tray>",
		Short: "Show or generate shareable invite for a tray",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return stub("invite")
		},
	}
}

func cmdJoin() *cobra.Command {
	return &cobra.Command{
		Use:   "join <url-or-token>",
		Short: "Join a tray via invite URL or token",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return stub("join")
		},
	}
}
