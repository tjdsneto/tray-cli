package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tjdsneto/tray-cli/internal/cli/trayref"
	"github.com/tjdsneto/tray-cli/internal/domain"
	"github.com/tjdsneto/tray-cli/internal/localtray"
	"github.com/tjdsneto/tray-cli/internal/output"
	"github.com/tjdsneto/tray-cli/internal/remotesfile"
)

func cmdAdd() *cobra.Command {
	c := &cobra.Command{
		Use:   `add "title" [tray]`,
		Short: "Add an item to a tray",
		Long: `Adds an item to a local or remote tray.

Local trays (no sign-in required): directory, branch, and named global trays stored on this machine.
Create-on-add: the tray is registered the first time you add an item unless --no-create is set.

Without a tray argument, adds to the branch tray on a non-default git branch, otherwise the current directory tray.

Remote trays (--remote, requires sign-in): accepted immediately on trays you own; pending when you contribute to someone else's tray.`,
		Args: cobra.RangeArgs(1, 2),
		RunE: runAdd,
	}
	c.Flags().Bool("here", false, "add to the directory tray for the current working directory")
	c.Flags().Bool("branch", false, "add to the branch tray for the current git branch")
	c.Flags().Bool("project", false, "add to the directory tray for the git repository root")
	c.Flags().Bool("remote", false, "add to a remote tray (requires sign-in)")
	c.Flags().Bool("no-create", false, "fail if the local tray does not exist yet")
	return c
}

func runAdd(cmd *cobra.Command, args []string) error {
	title := strings.TrimSpace(args[0])
	if title == "" {
		return fmt.Errorf("give the item a title — example: tray add \"Fix login\" inbox")
	}
	remote, err := cmd.Flags().GetBool("remote")
	if err != nil {
		return err
	}
	noCreate, err := cmd.Flags().GetBool("no-create")
	if err != nil {
		return err
	}
	create := !noCreate

	if remote {
		if len(args) < 2 {
			return fmt.Errorf("remote add requires a tray name — example: tray add \"Fix login\" inbox --remote")
		}
		return runRemoteItemAdd(cmd, title, strings.TrimSpace(args[1]))
	}

	cwd, git := workContext()
	here, _ := cmd.Flags().GetBool("here")
	branch, _ := cmd.Flags().GetBool("branch")
	project, _ := cmd.Flags().GetBool("project")

	var target localtray.AddTarget
	switch {
	case here || branch || project:
		target, err = localtray.TargetFromFlags(here, branch, project, cwd, git)
		if err != nil {
			return err
		}
	case len(args) == 2:
		target, err = localtray.TargetFromRef(strings.TrimSpace(args[1]), cwd, git)
		if err != nil {
			return err
		}
	default:
		target, err = localtray.DefaultAddTarget(cwd, git)
		if err != nil {
			return err
		}
	}
	return runLocalAdd(cmd, title, target, create)
}

func runRemoteItemAdd(cmd *cobra.Command, title, trayRef string) error {
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
	c := &cobra.Command{
		Use:   "list [tray]",
		Short: "List tray items",
		Long: `Default: open items on local trays in scope (global trays plus directory/branch trays registered along the path from the current working directory upward).

Use --remote for remote trays you own (requires sign-in). Use --all for every local tray ever created on this machine.

With a tray argument, lists that local tray by name or id. For remote trays, add --remote.`,
		Args: cobra.RangeArgs(0, 1),
		RunE: runList,
	}
	c.Flags().Bool("all", false, "list items on every local tray ever created")
	c.Flags().Bool("remote", false, "list items on remote trays you own (requires sign-in)")
	return c
}

func runList(cmd *cobra.Command, args []string) error {
	all, err := cmd.Flags().GetBool("all")
	if err != nil {
		return err
	}
	remote, err := cmd.Flags().GetBool("remote")
	if err != nil {
		return err
	}
	trayArg := ""
	if len(args) == 1 {
		trayArg = strings.TrimSpace(args[0])
	}
	if remote {
		return runRemoteItemList(cmd, trayArg)
	}
	cwd, git := workContext()
	ids, err := localListTrayIDs(all, cwd, git, trayArg)
	if err != nil {
		return err
	}
	return runLocalList(cmd, ids)
}

func runRemoteItemList(cmd *cobra.Command, trayArg string) error {
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
	if trayArg != "" {
		tid, err := trayref.TrayIDFromRef(trayArg, cmdDeps.RemoteAliases(), owned)
		if err != nil {
			return err
		}
		if _, ok := ownedIDs[strings.TrimSpace(tid)]; !ok {
			return fmt.Errorf("tray list --remote only shows trays you own — %q is not yours", trayArg)
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
	if trayArg == "" && len(owned) == 0 {
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
