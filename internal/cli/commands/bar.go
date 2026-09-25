package commands

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/spf13/cobra"
	"github.com/tjdsneto/tray-cli/internal/bar"
	"github.com/tjdsneto/tray-cli/internal/cli/trayref"
	"github.com/tjdsneto/tray-cli/internal/domain"
	"github.com/tjdsneto/tray-cli/internal/localtray"
)

func cmdBar() *cobra.Command {
	c := &cobra.Command{
		Use:   "bar",
		Short: "Show a macOS menu bar icon for open/pending tray items",
		RunE:  runBar,
	}
	c.Flags().Duration("interval", 30*time.Second, "refresh interval")
	c.Flags().Bool("daemon", false, "daemon mode: write bar.pid and bar.log under config dir")
	return c
}

func runBar(cmd *cobra.Command, args []string) error {
	return runBarUI(cmd)
}

// barRuntime holds refresh settings and the last successful snapshot.
type barRuntime struct {
	interval  time.Duration
	configDir string
	last      atomic.Pointer[bar.Snapshot]
}

func (r *barRuntime) lastOrEmpty() bar.Snapshot {
	if p := r.last.Load(); p != nil {
		return *p
	}
	return bar.Snapshot{}
}

func (r *barRuntime) store(snap bar.Snapshot) {
	cp := snap
	r.last.Store(&cp)
}

// keepLastRemoteUnavailable returns the previous snapshot with RemoteStatus set,
// or builds a local-only snapshot when there is no previous snapshot.
func (r *barRuntime) keepLastRemoteUnavailable(localItems []bar.Item) bar.Snapshot {
	if r.last.Load() != nil {
		snap := r.lastOrEmpty()
		snap.RemoteStatus = "remote unavailable"
		r.store(snap)
		return snap
	}
	snap := bar.BuildSnapshot(localItems, "remote unavailable")
	r.store(snap)
	return snap
}

// refresh loads local open items and, when authenticated, remote pending items.
// Auth failure → local only (no RemoteStatus). Remote I/O failure → keep last snapshot
// with "remote unavailable", or local-only if there is no previous snapshot.
func (r *barRuntime) refresh(cmd *cobra.Command) bar.Snapshot {
	ctx := cmd.Context()
	localRows, err := localtray.NewStore(r.configDir).AllOpenItems()
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "tray bar: local items: %v\n", err)
		return r.lastOrEmpty()
	}
	items := bar.ItemsFromLocal(localRows)

	svcs, sess, authErr := cmdDeps.RequireAuth()
	if authErr != nil {
		snap := bar.BuildSnapshot(items, "")
		r.store(snap)
		return snap
	}

	owned, oerr := svcs.Trays.ListOwned(ctx, sess)
	if oerr != nil {
		return r.keepLastRemoteUnavailable(items)
	}

	q := domain.ListItemsQuery{Status: "pending", OrderCreated: "desc"}
	if len(owned) == 0 {
		q.TrayID = noOwnedTraysTrayFilter
	} else {
		q.TrayIDIn = make([]string, 0, len(owned))
		for i := range owned {
			q.TrayIDIn = append(q.TrayIDIn, strings.TrimSpace(owned[i].ID))
		}
	}

	list, lerr := svcs.Items.List(ctx, sess, q)
	if lerr != nil {
		return r.keepLastRemoteUnavailable(items)
	}

	names := trayref.TrayNameMap(owned)
	items = append(items, bar.ItemsFromRemote(list, names)...)
	snap := bar.BuildSnapshot(items, "")
	r.store(snap)
	return snap
}
