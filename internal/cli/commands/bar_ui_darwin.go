//go:build darwin

package commands

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"fyne.io/systray"
	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
	"github.com/tjdsneto/tray-cli/internal/bar"
)

// Pool sized for ~40 rows plus section headers.
const barMenuPoolSize = 60

type barMenuSlot struct {
	item *systray.MenuItem
	row  bar.Row // set when slot is a clickable row; empty for headers
}

func runBarUI(cmd *cobra.Command) error {
	interval, err := cmd.Flags().GetDuration("interval")
	if err != nil {
		return err
	}
	if interval <= 0 {
		return fmt.Errorf("--interval must be greater than 0")
	}
	daemon, _ := cmd.Flags().GetBool("daemon")
	configDir := cmdDeps.ConfigDir()
	if daemon {
		cleanupPID, err := acquireBarDaemon(configDir)
		if err != nil {
			return err
		}
		defer cleanupPID()
		cleanupLog, err := redirectBarDaemonLog(cmd, configDir)
		if err != nil {
			return err
		}
		defer cleanupLog()
	}
	rt := &barRuntime{interval: interval, configDir: configDir}

	systray.Run(func() {
		systray.SetTitle("tray")
		systray.SetTooltip("tray")

		mRemote := systray.AddMenuItem("", "")
		mRemote.Disable()
		mRemote.Hide()

		slots := make([]barMenuSlot, barMenuPoolSize)
		var slotMu sync.Mutex
		for i := range slots {
			mi := systray.AddMenuItem("", "")
			mi.Hide()
			slots[i].item = mi
			idx := i
			go func() {
				for range mi.ClickedCh {
					slotMu.Lock()
					r := slots[idx].row
					slotMu.Unlock()
					if r.ItemID == "" {
						continue
					}
					if err := clipboard.WriteAll(bar.ClipboardLine(r)); err != nil {
						fmt.Fprintf(cmd.ErrOrStderr(), "tray bar: clipboard: %v\n", err)
					}
				}
			}()
		}

		systray.AddSeparator()
		mRefresh := systray.AddMenuItem("Refresh", "")
		mQuit := systray.AddMenuItem("Quit", "")

		apply := func(snap bar.Snapshot) {
			if snap.Badge > 0 {
				systray.SetTitle(fmt.Sprintf("tray %d", snap.Badge))
			} else {
				systray.SetTitle("tray")
			}

			if snap.RemoteStatus != "" {
				mRemote.SetTitle(snap.RemoteStatus)
				mRemote.Show()
			} else {
				mRemote.Hide()
			}

			slotMu.Lock()
			defer slotMu.Unlock()

			i := 0
			overflow := false
			for _, sec := range snap.Sections {
				if i >= len(slots) {
					overflow = true
					break
				}
				slots[i].item.SetTitle(sec.Header)
				slots[i].item.Disable()
				slots[i].item.Show()
				slots[i].row = bar.Row{}
				i++
				for _, row := range sec.Rows {
					if i >= len(slots) {
						overflow = true
						break
					}
					title := row.Title
					if title == "" {
						title = row.ItemID
					}
					slots[i].item.SetTitle(title)
					slots[i].item.Enable()
					slots[i].item.Show()
					slots[i].row = row
					i++
				}
			}
			for ; i < len(slots); i++ {
				slots[i].item.Hide()
				slots[i].row = bar.Row{}
			}
			if overflow {
				fmt.Fprintf(cmd.ErrOrStderr(), "tray bar: menu pool full (%d slots); some items hidden\n", len(slots))
			}
		}

		snapCh := make(chan bar.Snapshot, 1)
		var refreshing atomic.Bool
		kickRefresh := func() {
			if !refreshing.CompareAndSwap(false, true) {
				return
			}
			go func() {
				defer refreshing.Store(false)
				snap := rt.refresh(cmd)
				select {
				case snapCh <- snap:
				case <-cmd.Context().Done():
				}
			}()
		}

		kickRefresh()

		go func() {
			t := time.NewTicker(rt.interval)
			defer t.Stop()
			for {
				select {
				case <-cmd.Context().Done():
					systray.Quit()
					return
				case <-t.C:
					kickRefresh()
				case <-mRefresh.ClickedCh:
					kickRefresh()
				case snap := <-snapCh:
					apply(snap)
				case <-mQuit.ClickedCh:
					systray.Quit()
					return
				}
			}
		}()
	}, func() {})
	return nil
}
