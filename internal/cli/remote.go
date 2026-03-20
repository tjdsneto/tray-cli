package cli

import "github.com/spf13/cobra"

func cmdRemote() *cobra.Command {
	root := &cobra.Command{
		Use:   "remote",
		Short: "Manage local aliases for others' trays",
	}

	root.AddCommand(
		&cobra.Command{
			Use:   "add <alias> <invite-url-or-token>",
			Short: "Add a remote alias",
			Args:  cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				return stub("remote add")
			},
		},
		&cobra.Command{
			Use:   "ls",
			Short: "List remotes",
			RunE: func(cmd *cobra.Command, args []string) error {
				return stub("remote ls")
			},
		},
		&cobra.Command{
			Use:   "remove <alias>",
			Short: "Remove a remote alias",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				return stub("remote remove")
			},
		},
	)

	return root
}
