package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tjdsneto/tray-cli/internal/cli/trayref"
	"github.com/tjdsneto/tray-cli/internal/domain"
	"github.com/tjdsneto/tray-cli/internal/output"
	"github.com/tjdsneto/tray-cli/internal/remotesfile"
)

func cmdAdd() *cobra.Command {
	return &cobra.Command{
		Use:   `add "title" <tray>`,
		Short: "Add an item to a tray",
		Long:  `Adds an item: accepted immediately on trays you own; pending when you contribute to someone else's tray (they triage). Tray can be a name from tray ls, a joined tray (see tray remote ls), a remote alias, or a tray id.`,
		Args:  cobra.ExactArgs(2),
		RunE:  runAdd,
	}
}

func runAdd(cmd *cobra.Command, args []string) error {
	title := strings.TrimSpace(args[0])
	if title == "" {
		return fmt.Errorf("give the item a title — example: tray add \"Fix login\" inbox")
	}
	trayRef := strings.TrimSpace(args[1])
	aliases := cmdDeps.RemoteAliases()
	svcs, sess, err := cmdDeps.RequireAuth()
	if err != nil {
		return err
	}
	tid, err := trayref.ResolveTrayRef(cmd.Context(), svcs, sess, trayRef, aliases)
	if err != nil {
		return err
	}
	item, err := svcs.Items.Add(cmd.Context(), sess, tid, title, nil)
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
	m = withAddAliasDisplay(m, trayRef, tid, aliases)
	by := profileDisplayMap(cmd.Context(), sess, svcs, sourceUserIDsFromItems([]domain.Item{*item}))
	return output.WriteItems(cmd.OutOrStdout(), []domain.Item{*item}, m, strings.TrimSpace(sess.UserID), by, format)
}

func withAddAliasDisplay(trayNames map[string]string, trayRef, trayID string, aliases map[string]string) map[string]string {
	ref := strings.TrimSpace(trayRef)
	if ref == "" {
		return trayNames
	}
	if aliases == nil {
		return trayNames
	}
	if strings.TrimSpace(aliases[strings.ToLower(ref)]) != strings.TrimSpace(trayID) {
		return trayNames
	}
	out := make(map[string]string, len(trayNames)+1)
	for k, v := range trayNames {
		out[k] = v
	}
	out[strings.TrimSpace(trayID)] = ref
	return out
}

func cmdList() *cobra.Command {
	return &cobra.Command{
		Use:   "list [tray]",
		Short: "List items on trays you own (default: all of them)",
		Long: `Without arguments, lists items on every tray you own.

With a tray name, id, or remote alias, the tray must be one you own.
Items you filed on someone else's tray are listed with: tray contributed`,
		Args: cobra.RangeArgs(0, 1),
		RunE:  runList,
	}
}

func runList(cmd *cobra.Command, args []string) error {
	svcs, sess, err := cmdDeps.RequireAuth()
	if err != nil {
		return err
	}
	owned, err := svcs.Trays.ListOwned(cmd.Context(), sess)
	if err != nil {
		return err
	}
	ownedIDs := make(map[string]struct{}, len(owned))
	for i := range owned {
		ownedIDs[strings.TrimSpace(owned[i].ID)] = struct{}{}
	}
	q := domain.ListItemsQuery{}
	if len(args) == 1 {
		tid, err := trayref.ResolveTrayRef(cmd.Context(), svcs, sess, strings.TrimSpace(args[0]), cmdDeps.RemoteAliases())
		if err != nil {
			return err
		}
		if _, ok := ownedIDs[strings.TrimSpace(tid)]; !ok {
			return fmt.Errorf("tray list only shows trays you own — %q is not yours (items you added elsewhere: `tray contributed`; trays you joined: `tray remote ls`)", strings.TrimSpace(args[0]))
		}
		q.TrayID = tid
	} else {
		if len(owned) == 0 {
			q.TrayIDIn = nil
		} else {
			q.TrayIDIn = make([]string, 0, len(owned))
			for i := range owned {
				q.TrayIDIn = append(q.TrayIDIn, strings.TrimSpace(owned[i].ID))
			}
		}
	}
	var items []domain.Item
	if len(args) == 0 && len(owned) == 0 {
		items = nil
	} else {
		items, err = svcs.Items.List(cmd.Context(), sess, q)
		if err != nil {
			return err
		}
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
	return output.WriteItems(cmd.OutOrStdout(), items, m, strings.TrimSpace(sess.UserID), by, format)
}

func cmdContributed() *cobra.Command {
	return &cobra.Command{
		Use:   "contributed",
		Short: "List items you added to others' trays",
		RunE:  runContributed,
	}
}

func runContributed(cmd *cobra.Command, args []string) error {
	svcs, sess, err := cmdDeps.RequireAuth()
	if err != nil {
		return err
	}
	items, err := svcs.Items.ListOutbox(cmd.Context(), sess)
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
	if f, err := remotesfile.Load(cmdDeps.ConfigDir()); err == nil {
		m = trayref.OverlayTrayAliases(m, f.Aliases)
	}
	by := profileDisplayMap(cmd.Context(), sess, svcs, sourceUserIDsFromItems(items))
	return output.WriteItems(cmd.OutOrStdout(), items, m, strings.TrimSpace(sess.UserID), by, format)
}
