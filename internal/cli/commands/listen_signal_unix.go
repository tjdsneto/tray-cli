//go:build unix

package commands

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// notifySIGHUP calls onReload when the process receives SIGHUP (ignored on non-Unix builds without this file).
func notifySIGHUP(ctx context.Context, onReload func()) {
	if onReload == nil {
		return
	}
	ch := make(chan os.Signal, 2)
	signal.Notify(ch, syscall.SIGHUP)
	go func() {
		defer signal.Stop(ch)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ch:
				onReload()
			}
		}
	}()
}
