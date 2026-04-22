package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func cmdRemove() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <item-id>",
		Short: "Remove an item (tray owner: any item; contributor: pending only)",
		Long:  `Item id: full uuid from tray review / list / contributed, or a unique hex prefix (at least 8 characters) among items you own or filed on others' trays.`,
		Args:  cobra.ExactArgs(1),
		RunE:  runRemove,
	}
}

func runRemove(cmd *cobra.Command, args []string) error {
	svcs, sess, err := cmdDeps.RequireAuth()
	if err != nil {
		return err
	}
	id, err := resolveItemIDArg(cmd.Context(), svcs, sess, args[0], poolRemoveCandidates)
	if err != nil {
		return err
	}
	if err := svcs.Items.Delete(cmd.Context(), sess, id); err != nil {
		return err
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Removed item %s.\n", id)
	return err
}
