package cli

import "github.com/spf13/cobra"

func cmdAdd() *cobra.Command {
	return &cobra.Command{
		Use:   `add "title" [tray-or-alias]`,
		Short: "Add an item to a tray",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return stub("add")
		},
	}
}

func cmdList() *cobra.Command {
	return &cobra.Command{
		Use:   "list [tray]",
		Short: "List items in a tray (or all trays)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return stub("list")
		},
	}
}

func cmdContributed() *cobra.Command {
	return &cobra.Command{
		Use:   "contributed",
		Short: "List items you added to others' trays",
		RunE: func(cmd *cobra.Command, args []string) error {
			return stub("contributed")
		},
	}
}
