package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tjdsneto/tray-cli/internal/cli/trayref"
	"github.com/tjdsneto/tray-cli/internal/domain"
	"github.com/tjdsneto/tray-cli/internal/output"
)

func cmdListen() *cobra.Command {
	c := &cobra.Command{
		Use:   "listen [tray]",
		Short: "Watch for new pending items",
		Long:  `Polls for new pending items and prints updates as they arrive. Optionally scope to one tray by name/id/alias.`,
		Args:  cobra.RangeArgs(0, 1),
		RunE:  runListen,
	}
	c.Flags().Duration("interval", 10*time.Second, "poll interval (e.g. 5s, 30s)")
	c.Flags().Bool("once", false, "check once and exit")
	return c
}

func runListen(cmd *cobra.Command, args []string) error {
	svcs, sess, err := cmdDeps.RequireAuth()
	if err != nil {
		return err
	}
	interval, err := cmd.Flags().GetDuration("interval")
	if err != nil {
		return err
	}
	if interval <= 0 {
		return fmt.Errorf("--interval must be > 0")
	}
	once, err := cmd.Flags().GetBool("once")
	if err != nil {
		return err
	}

	q := domain.ListItemsQuery{Status: "pending", OrderCreated: "desc"}
	if len(args) == 1 {
		tid, err := trayref.ResolveTrayRef(cmd.Context(), svcs, sess, strings.TrimSpace(args[0]), cmdDeps.RemoteAliases())
		if err != nil {
			return err
		}
		q.TrayID = tid
	}

	format, err := output.FormatFromCmd(cmd)
	if err != nil {
		return err
	}
	trays, err := svcs.Trays.ListMine(cmd.Context(), sess)
	if err != nil {
		return err
	}
	trayNames := trayref.TrayNameMap(trays)

	items, err := svcs.Items.List(cmd.Context(), sess, q)
	if err != nil {
		return err
	}
	seen := make(map[string]struct{}, len(items))
	for _, it := range items {
		if strings.TrimSpace(it.ID) != "" {
			seen[it.ID] = struct{}{}
		}
	}

	if once {
		by := profileDisplayMap(cmd.Context(), sess, svcs, sourceUserIDsFromItems(items))
		return output.WriteItems(cmd.OutOrStdout(), items, trayNames, strings.TrimSpace(sess.UserID), by, format)
	}
	if format == output.FormatTable {
		if len(args) == 1 {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Listening for pending items in %q every %s...\n", strings.TrimSpace(args[0]), interval)
		} else {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Listening for pending items every %s...\n", interval)
		}
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-cmd.Context().Done():
			return nil
		case <-ticker.C:
			latest, err := svcs.Items.List(cmd.Context(), sess, q)
			if err != nil {
				return err
			}
			newItems := unseenItems(latest, seen)
			if len(newItems) > 0 {
				by := profileDisplayMap(cmd.Context(), sess, svcs, sourceUserIDsFromItems(newItems))
				if err := output.WriteItems(cmd.OutOrStdout(), newItems, trayNames, strings.TrimSpace(sess.UserID), by, format); err != nil {
					return err
				}
			}
		}
	}
}

func unseenItems(items []domain.Item, seen map[string]struct{}) []domain.Item {
	if seen == nil {
		seen = map[string]struct{}{}
	}
	out := make([]domain.Item, 0)
	for _, it := range items {
		id := strings.TrimSpace(it.ID)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, it)
	}
	return out
}
