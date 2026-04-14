package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func cmdItem() *cobra.Command {
	c := &cobra.Command{
		Use:   "item",
		Short: "Act on a single line item",
	}
	c.AddCommand(cmdItemUp(), cmdItemDown())
	return c
}

func cmdItemUp() *cobra.Command {
	return &cobra.Command{
		Use:   "up <item-id>",
		Short: "Move an item up in the tray list (owner only)",
			Long:  `Swaps manual order with the item above it. Order is the same as in tray list (sort_order).`,
		Args:  cobra.ExactArgs(1),
		RunE:  runItemUp,
	}
}

func cmdItemDown() *cobra.Command {
	return &cobra.Command{
		Use:   "down <item-id>",
		Short: "Move an item down in the tray list (owner only)",
		Long:  `Swaps manual order with the item below it. Order is the same as in tray list (sort_order).`,
		Args:  cobra.ExactArgs(1),
		RunE:  runItemDown,
	}
}

func runItemUp(cmd *cobra.Command, args []string) error {
	return runItemMove(cmd, args[0], -1)
}

func runItemDown(cmd *cobra.Command, args []string) error {
	return runItemMove(cmd, args[0], 1)
}

func runItemMove(cmd *cobra.Command, itemID string, dir int) error {
	id := strings.TrimSpace(itemID)
	if id == "" {
		return fmt.Errorf("pass the item id from tray list (see column id with --format json)")
	}
	svcs, sess, err := cmdDeps.RequireAuth()
	if err != nil {
		return err
	}
	var moveErr error
	if dir < 0 {
		moveErr = svcs.Items.MoveUp(cmd.Context(), sess, id)
	} else {
		moveErr = svcs.Items.MoveDown(cmd.Context(), sess, id)
	}
	if moveErr != nil {
		return moveErr
	}
	verb := "up"
	if dir > 0 {
		verb = "down"
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Moved item %s %s.\n", id, verb)
	return err
}
