package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

const (
	listenPidFile = "listen.pid"
	listenLogFile = "listen.log"
)

func acquireListenDaemon(configDir string) (cleanup func(), err error) {
	pidPath := filepath.Join(configDir, listenPidFile)
	if b, err := os.ReadFile(pidPath); err == nil {
		old, _ := strconv.Atoi(strings.TrimSpace(string(b)))
		if old > 0 && pidAlive(old) {
			return nil, fmt.Errorf("listen daemon already running (pid %d)", old)
		}
		_ = os.Remove(pidPath)
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	f, err := os.OpenFile(pidPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return nil, err
	}
	_, werr := fmt.Fprintf(f, "%d\n", os.Getpid())
	if cerr := f.Close(); werr != nil || cerr != nil {
		_ = os.Remove(pidPath)
		if werr != nil {
			return nil, werr
		}
		return nil, cerr
	}

	return func() { _ = os.Remove(pidPath) }, nil
}

func redirectListenDaemonLog(cmd *cobra.Command, configDir string) (cleanup func(), err error) {
	logPath := filepath.Join(configDir, listenLogFile)
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return nil, err
	}
	cmd.SetOut(f)
	cmd.SetErr(f)
	return func() { _ = f.Close() }, nil
}

// hooksWatch describes a single file (by directory + base name) to watch for reload.
type hooksWatch struct {
	dir  string
	base string
}

func startHooksReloadWatch(ctx context.Context, w hooksWatch, reload func()) {
	if reload == nil || strings.TrimSpace(w.dir) == "" || strings.TrimSpace(w.base) == "" {
		return
	}
	dir, base := w.dir, w.base
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "listen: hooks watch: %v\n", err)
		return
	}
	if err := watcher.Add(dir); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "listen: hooks watch %q: %v\n", dir, err)
		_ = watcher.Close()
		return
	}
	var mu sync.Mutex
	var timer *time.Timer
	debounce := 300 * time.Millisecond
	schedule := func() {
		mu.Lock()
		defer mu.Unlock()
		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(debounce, func() {
			reload()
		})
	}
	go func() {
		defer func() {
			mu.Lock()
			if timer != nil {
				timer.Stop()
			}
			mu.Unlock()
			_ = watcher.Close()
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case ev, ok := <-watcher.Events:
				if !ok {
					return
				}
				if ev.Has(fsnotify.Create) || ev.Has(fsnotify.Write) || ev.Has(fsnotify.Rename) || ev.Has(fsnotify.Remove) {
					if nameMatchesWatch(ev.Name, dir, base) {
						schedule()
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				if err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "listen: hooks watch: %v\n", err)
				}
			}
		}
	}()
}

func nameMatchesWatch(fullPath, dir, base string) bool {
	fullPath = filepath.Clean(fullPath)
	if filepath.Base(fullPath) != base {
		return false
	}
	if filepath.Clean(filepath.Dir(fullPath)) != filepath.Clean(dir) {
		return false
	}
	return true
}
