package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func cmdRemove() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <item-id>",
		Short: "Remove an item (tray owner: any item; contributor: pending only)",
		Args:  cobra.ExactArgs(1),
		RunE:  runRemove,
	}
}

func runRemove(cmd *cobra.Command, args []string) error {
	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("pass the item id from `tray list --format json`")
	}
	svcs, sess, err := cmdDeps.RequireAuth()
	if err != nil {
		return err
	}
	if err := svcs.Items.Delete(cmd.Context(), sess, id); err != nil {
		return err
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Removed item %s.\n", id)
	return err
}
