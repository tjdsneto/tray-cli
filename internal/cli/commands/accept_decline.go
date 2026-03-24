package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tjdsneto/tray-cli/internal/domain"
)

func cmdAccept() *cobra.Command {
	return &cobra.Command{
		Use:   "accept <item-id>",
		Short: "Accept a pending item (tray owner)",
		Long:  `Sets the item status to "accepted". Use the item id from tray list --format json.`,
		Args:  cobra.ExactArgs(1),
		RunE:  runAccept,
	}
}

func runAccept(cmd *cobra.Command, args []string) error {
	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("pass the item id from `tray list --format json`")
	}
	svcs, sess, err := cmdDeps.RequireAuth()
	if err != nil {
		return err
	}
	st := "accepted"
	if err := svcs.Items.Update(cmd.Context(), sess, id, domain.ItemPatch{Status: &st}); err != nil {
		return err
	}
	return printTriageResult(cmd, svcs, sess, id, "Accepted")
}

func cmdDecline() *cobra.Command {
	c := &cobra.Command{
		Use:   "decline <item-id>",
		Short: "Decline a pending item (tray owner)",
		Long:  `Sets the item status to "declined". Optional --reason is stored for the contributor.`,
		Args:  cobra.ExactArgs(1),
		RunE:  runDecline,
	}
	c.Flags().String("reason", "", "optional note for the person who added the item")
	return c
}

func runDecline(cmd *cobra.Command, args []string) error {
	id := strings.TrimSpace(args[0])
	if id == "" {
		return fmt.Errorf("pass the item id from `tray list --format json`")
	}
	svcs, sess, err := cmdDeps.RequireAuth()
	if err != nil {
		return err
	}
	reason, err := cmd.Flags().GetString("reason")
	if err != nil {
		return err
	}
	st := "declined"
	patch := domain.ItemPatch{Status: &st}
	if strings.TrimSpace(reason) != "" {
		r := strings.TrimSpace(reason)
		patch.DeclineReason = &r
	}
	if err := svcs.Items.Update(cmd.Context(), sess, id, patch); err != nil {
		return err
	}
	return printTriageResult(cmd, svcs, sess, id, "Declined")
}

func printTriageResult(cmd *cobra.Command, svcs domain.Services, sess domain.Session, itemID, verb string) error {
	items, err := svcs.Items.List(cmd.Context(), sess, domain.ListItemsQuery{ItemID: itemID})
	if err != nil || len(items) != 1 {
		_, err := fmt.Fprintf(cmd.OutOrStdout(), "%s item %s.\n", verb, itemID)
		return err
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s %q (%s).\n", verb, items[0].Title, itemID)
	return err
}
