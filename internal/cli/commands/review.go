package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tjdsneto/tray-cli/internal/cli/trayref"
	"github.com/tjdsneto/tray-cli/internal/output"
)

func cmdReview() *cobra.Command {
	return &cobra.Command{
		Use:   "review [tray]",
		Short: "Review pending items (owner triage queue)",
		Long:  `Lists pending items for triage. Optionally pass a tray name/id/alias to focus on one tray.`,
		Args:  cobra.RangeArgs(0, 1),
		RunE:  runReview,
	}
}

func runReview(cmd *cobra.Command, args []string) error {
	svcs, sess, err := cmdDeps.RequireAuth()
	if err != nil {
		return err
	}

	q, err := pendingItemsOnOwnedTraysQuery(cmd.Context(), svcs, sess, optionalTrayRefArg(args), cmdDeps.RemoteAliases())
	if err != nil {
		return err
	}
	items, err := svcs.Items.List(cmd.Context(), sess, q)
	if err != nil {
		return err
	}
	format, err := output.FormatFromCmd(cmd)
	if err != nil {
		return err
	}
	trays, err := svcs.Trays.ListMine(cmd.Context(), sess)
	if err != nil {
		return err
	}
	m := trayref.TrayNameMap(trays)
	by := profileDisplayMap(cmd.Context(), sess, svcs, sourceUserIDsFromItems(items))
	if err := output.WriteItems(cmd.OutOrStdout(), items, m, strings.TrimSpace(sess.UserID), by, format); err != nil {
		return err
	}
	if format == output.FormatTable && len(items) == 0 {
		_, err := fmt.Fprintln(cmd.OutOrStdout(), "Tip: items appear here when contributors add to your trays.")
		return err
	}
	return nil
}
