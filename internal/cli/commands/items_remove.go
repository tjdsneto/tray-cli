package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func cmdRemove() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <item-id>",
		Short: "Remove an item",
		Long:  `Removes a local item (from tray list) or a remote item (tray owner: any item; contributor: pending only). Item id: full id or a unique hex prefix (at least 8 characters) among open local items or remote items in scope.`,
		Args:  cobra.ExactArgs(1),
		RunE:  runRemove,
	}
}

func runRemove(cmd *cobra.Command, args []string) error {
	if id, ok, err := tryLocalRemove(args[0]); ok {
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Removed local item %s.\n", id)
		return err
	}
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
