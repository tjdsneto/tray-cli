package commands

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/spf13/cobra"
	"github.com/tjdsneto/tray-cli/internal/bar"
	"github.com/tjdsneto/tray-cli/internal/cli/trayref"
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

// refresh loads local open items and, when authenticated, remote pending items.
// Auth failure → local only (no RemoteStatus). Remote I/O failure → local + "remote unavailable".
func (r *barRuntime) refresh(cmd *cobra.Command) bar.Snapshot {
	ctx := cmd.Context()
	localRows, err := localtray.NewStore(r.configDir).AllOpenItems()
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "tray bar: local items: %v\n", err)
		return r.lastOrEmpty()
	}
	items := bar.ItemsFromLocal(localRows)
	remoteStatus := ""

	svcs, sess, authErr := cmdDeps.RequireAuth()
	if authErr == nil {
		aliases := cmdDeps.RemoteAliases()
		q, qerr := pendingItemsOnOwnedTraysQuery(ctx, svcs, sess, "", aliases)
		if qerr != nil {
			remoteStatus = "remote unavailable"
		} else {
			q.OrderCreated = "desc"
			list, lerr := svcs.Items.List(ctx, sess, q)
			if lerr != nil {
				remoteStatus = "remote unavailable"
			} else {
				owned, oerr := svcs.Trays.ListOwned(ctx, sess)
				if oerr != nil {
					remoteStatus = "remote unavailable"
				} else {
					names := trayref.TrayNameMap(owned)
					items = append(items, bar.ItemsFromRemote(list, names)...)
				}
			}
		}
	}

	snap := bar.BuildSnapshot(items, remoteStatus)
	r.store(snap)
	return snap
}
