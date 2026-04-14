package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/spf13/cobra"
	"github.com/tjdsneto/tray-cli/internal/cli/listenhook"
	"github.com/tjdsneto/tray-cli/internal/cli/trayref"
	"github.com/tjdsneto/tray-cli/internal/config"
	"github.com/tjdsneto/tray-cli/internal/domain"
	"github.com/tjdsneto/tray-cli/internal/output"
	"github.com/tjdsneto/tray-cli/internal/supabase"
)

func cmdListen() *cobra.Command {
	c := &cobra.Command{
		Use:   "listen [tray]",
		Short: "Watch for new items or run hooks from hooks.json",
		Long: `Polls the API and prints new pending items, and/or runs hooks defined in hooks.json (in-memory only; no local database).

Hooks file defaults to $TRAY_CONFIG_DIR/hooks.json when present (events: item.pending, item.completed, item.accepted, item.declined for outbox).
Hooks receive TRAY_HOOK_EVENT, TRAY_SESSION_USER_ID, TRAY_ITEM_*, and timestamps; TRAY_ITEM_DECLINE_REASON when the owner declined. See docs/user/hooks.md for recipes and env reference.

When --once is set, prints a one-shot snapshot of pending items and does not run hooks.`,
		Args: cobra.RangeArgs(0, 1),
		RunE: runListen,
	}
	c.Flags().Duration("interval", 10*time.Second, "poll interval (e.g. 5s, 30s)")
	c.Flags().Bool("once", false, "check once: print current pending items and exit (hooks disabled)")
	c.Flags().String("hooks", "", "path to hooks JSON (default: <config-dir>/hooks.json if that file exists)")
	c.Flags().Bool("no-hooks", false, "do not load hooks even if hooks.json exists")
	c.Flags().Bool("quiet", false, "do not print new items to stdout (hooks still run)")
	c.Flags().String("exec", "", "compat: run this shell command when matching events fire")
	c.Flags().String("exec-pattern", "item.pending", "compat: comma-separated events for --exec (pending,completed,accepted,declined or item.*)")
	c.Flags().Bool("verbose", false, "print detailed hook diagnostics (match/run/ok) to stderr")
	c.Flags().String("mode", "auto", "listen mode: auto (realtime then fallback), realtime (strict), or poll")
	c.Flags().Bool("daemon", false, "daemon mode: write listen.pid and listen.log under config dir, reload hooks on hooks.json change (and SIGHUP on Unix)")
	return c
}

func runListen(cmd *cobra.Command, args []string) error {
	svcs, sess, err := cmdDeps.RequireAuth()
	if err != nil {
		return err
	}
	daemon, err := cmd.Flags().GetBool("daemon")
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
	if daemon && once {
		return fmt.Errorf("--daemon cannot be used with --once")
	}
	quiet, err := cmd.Flags().GetBool("quiet")
	if err != nil {
		return err
	}
	modeRaw, err := cmd.Flags().GetString("mode")
	if err != nil {
		return err
	}
	mode := strings.ToLower(strings.TrimSpace(modeRaw))
	switch mode {
	case "auto", "realtime", "poll":
	default:
		return fmt.Errorf("--mode must be one of: auto, realtime, poll")
	}

	q := domain.ListItemsQuery{Status: "pending"}
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
	owned := listenhook.OwnedTraySet(trays, sess.UserID)

	cfg, ruleTray, hookPath, err := mergedListenHooks(cmd, svcs, sess)
	if err != nil {
		return err
	}
	if cfg != nil {
		if err := validateHookCoverage(cfg, hookPath); err != nil {
			return err
		}
	}

	var hooks atomic.Pointer[hookRuntime]
	if cfg != nil {
		hooks.Store(&hookRuntime{cfg: cfg, ruleTray: ruleTray, hookPathLabel: hookPath})
	}

	reloadHooks := func() {
		cfg2, rt2, lab2, err := mergedListenHooks(cmd, svcs, sess)
		if err != nil {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "listen: reload hooks: %v\n", err)
			return
		}
		if cfg2 == nil {
			hooks.Store(nil)
			return
		}
		if err := validateHookCoverage(cfg2, lab2); err != nil {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "listen: reload hooks: %v\n", err)
			return
		}
		hooks.Store(&hookRuntime{cfg: cfg2, ruleTray: rt2, hookPathLabel: lab2})
	}

	if daemon {
		cleanupPID, err := acquireListenDaemon(cmdDeps.ConfigDir())
		if err != nil {
			return err
		}
		defer cleanupPID()
		cleanupLog, err := redirectListenDaemonLog(cmd, cmdDeps.ConfigDir())
		if err != nil {
			return err
		}
		defer cleanupLog()
		startHooksReloadWatch(cmd.Context(), hooksWatchPath(cmd), reloadHooks)
		notifySIGHUP(cmd.Context(), reloadHooks)
	}

	hookCfg, _, _ := snapHooks(&hooks)
	needPending := hookCfg == nil || hookCfg.WantsPendingPoll() || format == output.FormatTable

	var items []domain.Item
	if needPending || once {
		var err error
		items, sess, err = listItemsWithRetry(cmd.Context(), svcs, sess, q)
		if err != nil {
			return err
		}
	}

	if once {
		by := profileDisplayMap(cmd.Context(), sess, svcs, sourceUserIDsFromItems(items))
		return output.WriteItems(cmd.OutOrStdout(), items, trayNames, strings.TrimSpace(sess.UserID), by, format)
	}

	if mode != "poll" {
		rtErr := runListenRealtime(cmd, svcs, sess, q, format, quiet, &hooks, trayNames, owned)
		if rtErr == nil {
			return nil
		}
		if mode == "realtime" {
			return rtErr
		}
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "realtime unavailable (%v); falling back to polling (use --mode poll to skip this step)\n", rtErr)
	}

	return runListenPoll(cmd, svcs, sess, q, format, quiet, &hooks, trayNames, owned, interval, items)
}

func runListenPoll(cmd *cobra.Command, svcs domain.Services, sess domain.Session, q domain.ListItemsQuery, format output.Format, quiet bool, hooks *atomic.Pointer[hookRuntime], trayNames map[string]string, owned map[string]bool, interval time.Duration, seedItems []domain.Item) error {
	hookCfg0, _, _ := snapHooks(hooks)
	needPending0 := hookCfg0 == nil || hookCfg0.WantsPendingPoll() || format == output.FormatTable
	needOutbox0 := hookCfg0 != nil && hookCfg0.WantsOutboxPoll()
	pendingSeen := listenhook.NewPendingSeen()
	outboxState := listenhook.NewOutboxState()

	if needPending0 {
		pendingSeen.Seed(seedItems)
	}
	var outboxSeeded bool
	if needOutbox0 {
		outboxItems, s, err := listOutboxWithRetry(cmd.Context(), svcs, sess)
		if err != nil {
			return err
		}
		sess = s
		outboxState.Seed(outboxItems)
		outboxSeeded = true
	}

	if format == output.FormatTable && !quiet {
		hookCfg, _, hookPath := snapHooks(hooks)
		needPending := hookCfg == nil || hookCfg.WantsPendingPoll() || format == output.FormatTable
		needOutbox := hookCfg != nil && hookCfg.WantsOutboxPoll()
		switch {
		case hookPath != "" && needPending && needOutbox:
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Listening (pending + outbox hooks) %q every %s...\n", hookPath, interval)
		case hookPath != "" && needOutbox:
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Listening (outbox hooks) %q every %s...\n", hookPath, interval)
		case hookPath != "":
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Listening with hooks %q every %s...\n", hookPath, interval)
		case strings.TrimSpace(q.TrayID) != "":
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Listening for pending items in selected tray every %s...\n", interval)
		default:
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Listening for pending items every %s...\n", interval)
		}
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	runTick := func() error {
		hookCfg, ruleTray, _ := snapHooks(hooks)
		needPending := hookCfg == nil || hookCfg.WantsPendingPoll() || format == output.FormatTable
		needOutbox := hookCfg != nil && hookCfg.WantsOutboxPoll()
		if needOutbox {
			if !outboxSeeded {
				outboxItems, s, err := listOutboxWithRetry(cmd.Context(), svcs, sess)
				if err != nil {
					return err
				}
				sess = s
				outboxState.Seed(outboxItems)
				outboxSeeded = true
			}
		} else {
			outboxSeeded = false
		}
		if needPending {
			latest, s, err := listItemsWithRetry(cmd.Context(), svcs, sess, q)
			if err != nil {
				return err
			}
			sess = s
			newItems := pendingSeen.NewPending(latest)
			if format == output.FormatTable {
				by := profileDisplayMap(cmd.Context(), sess, svcs, sourceUserIDsFromItems(newItems))
				for _, it := range newItems {
					emitListenEvent(cmd, listenhook.EventItemPending, it, trayNames, strings.TrimSpace(sess.UserID), by)
				}
			}
			if len(newItems) > 0 && !quiet {
				by := profileDisplayMap(cmd.Context(), sess, svcs, sourceUserIDsFromItems(newItems))
				if err := output.WriteItems(cmd.OutOrStdout(), newItems, trayNames, strings.TrimSpace(sess.UserID), by, format); err != nil {
					return err
				}
			}
			if hookCfg != nil {
				by := profileDisplayMap(cmd.Context(), sess, svcs, sourceUserIDsFromItems(newItems))
				for _, it := range newItems {
					srcName := strings.TrimSpace(by[strings.TrimSpace(it.SourceUserID)])
					for i, r := range hookCfg.Hooks {
						if strings.TrimSpace(r.Event) != listenhook.EventItemPending {
							continue
						}
						tf := ""
						if i < len(ruleTray) {
							tf = ruleTray[i]
						}
						if !listenhook.MatchPending(r, it, owned, sess.UserID, tf) {
							continue
						}
						runHookWithLogs(cmd, r, listenhook.EventItemPending, sess, it, srcName)
					}
				}
			}
		}
		if needOutbox && hookCfg != nil {
			outLatest, s, err := listOutboxWithRetry(cmd.Context(), svcs, sess)
			if err != nil {
				return err
			}
			sess = s
			trans := outboxState.OutTransitions(outLatest)
			runOutboxHooks(cmd, svcs, sess, hookCfg, ruleTray, trans)
		}
		return nil
	}

	return runTickLoop(cmd, ticker, runTick)
}

func runOutboxHooks(cmd *cobra.Command, svcs domain.Services, sess domain.Session, hookCfg *listenhook.Config, ruleTray []string, trans listenhook.OutboxTransitions) {
	runOne := func(items []domain.Item, event string) {
		if len(items) == 0 {
			return
		}
		by := profileDisplayMap(cmd.Context(), sess, svcs, sourceUserIDsFromItems(items))
		trays, err := svcs.Trays.ListMine(cmd.Context(), sess)
		if err != nil {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "listen: unable to resolve tray names for event output: %v\n", err)
		}
		trayNames := trayref.TrayNameMap(trays)
		for _, it := range items {
			emitListenEvent(cmd, event, it, trayNames, strings.TrimSpace(sess.UserID), by)
		}
		for _, it := range items {
			srcName := strings.TrimSpace(by[strings.TrimSpace(it.SourceUserID)])
			for i, r := range hookCfg.Hooks {
				if strings.TrimSpace(r.Event) != event {
					continue
				}
				tf := ""
				if i < len(ruleTray) {
					tf = ruleTray[i]
				}
				if !listenhook.MatchOutboxFilter(r, it, tf) {
					continue
				}
				runHookWithLogs(cmd, r, event, sess, it, srcName)
			}
		}
	}
	runOne(trans.Completed, listenhook.EventItemCompleted)
	runOne(trans.Accepted, listenhook.EventItemAccepted)
	runOne(trans.Declined, listenhook.EventItemDeclined)
}

func runTickLoop(cmd *cobra.Command, ticker *time.Ticker, runTick func() error) error {
	if err := runTick(); err != nil {
		return err
	}
	for {
		select {
		case <-cmd.Context().Done():
			return nil
		case <-ticker.C:
			if err := runTick(); err != nil {
				return err
			}
		}
	}
}

func mergedListenHooks(cmd *cobra.Command, svcs domain.Services, sess domain.Session) (*listenhook.Config, []string, string, error) {
	hookCfg, hookPath, err := loadListenHookConfig(cmd)
	if err != nil {
		return nil, nil, "", err
	}
	execCommand, err := cmd.Flags().GetString("exec")
	if err != nil {
		return nil, nil, "", err
	}
	execPattern, err := cmd.Flags().GetString("exec-pattern")
	if err != nil {
		return nil, nil, "", err
	}
	if strings.TrimSpace(execCommand) != "" {
		rules, err := execRules(strings.TrimSpace(execCommand), strings.TrimSpace(execPattern))
		if err != nil {
			return nil, nil, "", err
		}
		if hookCfg == nil {
			hookCfg = &listenhook.Config{}
		}
		hookCfg.Hooks = append(hookCfg.Hooks, rules...)
		if hookPath == "" {
			hookPath = "--exec"
		} else {
			hookPath += " + --exec"
		}
	}
	if hookCfg == nil {
		return nil, nil, "", nil
	}
	ruleTray, err := resolveHookRuleTrays(cmd, svcs, sess, hookCfg.Hooks)
	if err != nil {
		return nil, nil, "", err
	}
	return hookCfg, ruleTray, hookPath, nil
}

func validateHookCoverage(cfg *listenhook.Config, label string) error {
	if cfg == nil {
		return nil
	}
	needPending := cfg.WantsPendingPoll()
	needOutbox := cfg.WantsOutboxPoll()
	if !needPending && !needOutbox {
		if label == "" {
			label = "hooks"
		}
		return fmt.Errorf("hooks file %q has no supported hook events", label)
	}
	return nil
}

func hooksWatchPath(cmd *cobra.Command) hooksWatch {
	noHooks, err := cmd.Flags().GetBool("no-hooks")
	if err != nil || noHooks {
		return hooksWatch{}
	}
	explicit, err := cmd.Flags().GetString("hooks")
	if err != nil {
		return hooksWatch{}
	}
	if strings.TrimSpace(explicit) != "" {
		p := filepath.Clean(strings.TrimSpace(explicit))
		return hooksWatch{dir: filepath.Dir(p), base: filepath.Base(p)}
	}
	p := filepath.Join(cmdDeps.ConfigDir(), "hooks.json")
	return hooksWatch{dir: filepath.Dir(p), base: filepath.Base(p)}
}

func loadListenHookConfig(cmd *cobra.Command) (*listenhook.Config, string, error) {
	noHooks, err := cmd.Flags().GetBool("no-hooks")
	if err != nil {
		return nil, "", err
	}
	if noHooks {
		return nil, "", nil
	}
	explicit, err := cmd.Flags().GetString("hooks")
	if err != nil {
		return nil, "", err
	}
	if strings.TrimSpace(explicit) != "" {
		path := strings.TrimSpace(explicit)
		if _, err := os.Stat(path); err != nil {
			return nil, "", err
		}
		cfg, err := listenhook.Load(path)
		return cfg, path, err
	}
	path := filepath.Join(cmdDeps.ConfigDir(), "hooks.json")
	if _, err := os.Stat(path); err != nil {
		return nil, "", nil
	}
	cfg, err := listenhook.Load(path)
	return cfg, path, err
}

func resolveHookRuleTrays(cmd *cobra.Command, svcs domain.Services, sess domain.Session, rules []listenhook.Rule) ([]string, error) {
	out := make([]string, len(rules))
	for i, r := range rules {
		ref := strings.TrimSpace(r.Tray)
		if ref == "" {
			continue
		}
		tid, err := trayref.ResolveTrayRef(cmd.Context(), svcs, sess, ref, cmdDeps.RemoteAliases())
		if err != nil {
			return nil, err
		}
		out[i] = tid
	}
	return out, nil
}

func execRules(execCommand, pattern string) ([]listenhook.Rule, error) {
	if strings.TrimSpace(execCommand) == "" {
		return nil, fmt.Errorf("--exec must not be empty")
	}
	events, err := parseExecEvents(pattern)
	if err != nil {
		return nil, err
	}
	cmd := shellCommand(execCommand)
	out := make([]listenhook.Rule, 0, len(events))
	for _, ev := range events {
		out = append(out, listenhook.Rule{
			Event:   ev,
			Command: cmd,
		})
	}
	return out, nil
}

func parseExecEvents(pattern string) ([]string, error) {
	raw := strings.TrimSpace(pattern)
	if raw == "" {
		raw = listenhook.EventItemPending
	}
	tokens := strings.Fields(strings.ReplaceAll(raw, ",", " "))
	seen := map[string]struct{}{}
	var out []string
	add := func(ev string) {
		if _, ok := seen[ev]; ok {
			return
		}
		seen[ev] = struct{}{}
		out = append(out, ev)
	}
	for _, t := range tokens {
		switch strings.ToLower(strings.TrimSpace(t)) {
		case "all":
			add(listenhook.EventItemPending)
			add(listenhook.EventItemCompleted)
			add(listenhook.EventItemAccepted)
			add(listenhook.EventItemDeclined)
		case "pending", listenhook.EventItemPending:
			add(listenhook.EventItemPending)
		case "completed", listenhook.EventItemCompleted:
			add(listenhook.EventItemCompleted)
		case "accepted", listenhook.EventItemAccepted:
			add(listenhook.EventItemAccepted)
		case "declined", listenhook.EventItemDeclined:
			add(listenhook.EventItemDeclined)
		default:
			return nil, fmt.Errorf("--exec-pattern: unsupported event %q", t)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("--exec-pattern must include at least one event")
	}
	return out, nil
}

func shellCommand(command string) []string {
	if runtime.GOOS == "windows" {
		return []string{"cmd", "/C", command}
	}
	return []string{"/bin/sh", "-c", command}
}

// Backoff between failed SubscribeItems attempts (3 waits => 1 initial + 3 retries before poll fallback).
var realtimeListenBackoff = []time.Duration{400 * time.Millisecond, 900 * time.Millisecond, 2 * time.Second}

func realtimeSubscribeBackoffDelay(failureCount int) time.Duration {
	i := failureCount - 1
	if i < 0 || len(realtimeListenBackoff) == 0 {
		return 0
	}
	if i >= len(realtimeListenBackoff) {
		return realtimeListenBackoff[len(realtimeListenBackoff)-1]
	}
	return realtimeListenBackoff[i]
}

func runListenRealtime(cmd *cobra.Command, svcs domain.Services, sess domain.Session, q domain.ListItemsQuery, format output.Format, quiet bool, hooks *atomic.Pointer[hookRuntime], trayNames map[string]string, owned map[string]bool) error {
	hookCfg, _, hookPath := snapHooks(hooks)
	needPending := hookCfg == nil || hookCfg.WantsPendingPoll() || format == output.FormatTable
	needOutbox := hookCfg != nil && hookCfg.WantsOutboxPoll()
	if !needPending && !needOutbox {
		return fmt.Errorf("no events requested")
	}

	maxSubscribeFails := 1 + len(realtimeListenBackoff)
	consecutiveSubscribeFails := 0
	subscribedOnce := false

	for {
		if err := cmd.Context().Err(); err != nil {
			return err
		}

		rc, err := supabase.NewRealtimeClient(config.SupabaseURL(), config.SupabaseAnonKey())
		if err != nil {
			return err
		}
		changes, errs, err := rc.SubscribeItems(cmd.Context(), sess.AccessToken)
		if err != nil {
			if !supabase.IsRetryableRealtimeErr(err) {
				return err
			}
			consecutiveSubscribeFails++
			if consecutiveSubscribeFails >= maxSubscribeFails {
				return err
			}
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "realtime unavailable (%v); retrying (%d/%d)...\n", err, consecutiveSubscribeFails, len(realtimeListenBackoff))
			select {
			case <-cmd.Context().Done():
				return cmd.Context().Err()
			case <-time.After(realtimeSubscribeBackoffDelay(consecutiveSubscribeFails)):
			}
			continue
		}

		consecutiveSubscribeFails = 0

		if !subscribedOnce {
			if format == output.FormatTable && !quiet {
				switch {
				case hookPath != "" && needPending && needOutbox:
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Listening realtime (pending + outbox hooks) %q...\n", hookPath)
				case hookPath != "" && needOutbox:
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Listening realtime (outbox hooks) %q...\n", hookPath)
				case hookPath != "":
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Listening realtime with hooks %q...\n", hookPath)
				default:
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Listening realtime for pending items...\n")
				}
			}
			subscribedOnce = true
		} else {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "realtime reconnected\n")
		}

		streamErr := runListenRealtimeDispatchLoop(cmd, svcs, sess, q, format, quiet, hooks, trayNames, owned, changes, errs)
		if err := cmd.Context().Err(); err != nil {
			return err
		}
		if streamErr == nil {
			return nil
		}
		if !supabase.IsRetryableRealtimeErr(streamErr) {
			return streamErr
		}
		// Transient disconnect: reconnect without counting toward subscribe-fail limit.
	}
}

func runListenRealtimeDispatchLoop(cmd *cobra.Command, svcs domain.Services, sess domain.Session, q domain.ListItemsQuery, format output.Format, quiet bool, hooks *atomic.Pointer[hookRuntime], trayNames map[string]string, owned map[string]bool, changes <-chan supabase.Change, errs <-chan error) error {
	for {
		select {
		case <-cmd.Context().Done():
			return nil
		case err := <-errs:
			if err == nil {
				return fmt.Errorf("realtime disconnected")
			}
			return err
		case ch, ok := <-changes:
			if !ok {
				return fmt.Errorf("realtime stream closed")
			}
			if err := handleRealtimeChange(cmd.Context(), cmd, svcs, sess, q, format, quiet, hooks, trayNames, owned, ch); err != nil {
				return err
			}
		}
	}
}

func handleRealtimeChange(ctx context.Context, cmd *cobra.Command, svcs domain.Services, sess domain.Session, q domain.ListItemsQuery, format output.Format, quiet bool, hooks *atomic.Pointer[hookRuntime], trayNames map[string]string, owned map[string]bool, ch supabase.Change) error {
	hookCfg, ruleTray, _ := snapHooks(hooks)
	needPending := hookCfg == nil || hookCfg.WantsPendingPoll() || format == output.FormatTable
	needOutbox := hookCfg != nil && hookCfg.WantsOutboxPoll()
	newItem, hasNew := itemFromRealtime(ch.New)
	oldItem, hasOld := itemFromRealtime(ch.Old)
	if hasNew && strings.TrimSpace(q.TrayID) != "" && strings.TrimSpace(newItem.TrayID) != strings.TrimSpace(q.TrayID) {
		return nil
	}
	if hasOld && strings.TrimSpace(q.TrayID) != "" && strings.TrimSpace(oldItem.TrayID) != strings.TrimSpace(q.TrayID) {
		return nil
	}

	if needPending && strings.EqualFold(ch.Type, "INSERT") && hasNew && strings.EqualFold(strings.TrimSpace(newItem.Status), "pending") {
		by := profileDisplayMap(ctx, sess, svcs, sourceUserIDsFromItems([]domain.Item{newItem}))
		if format == output.FormatTable {
			emitListenEvent(cmd, listenhook.EventItemPending, newItem, trayNames, strings.TrimSpace(sess.UserID), by)
		}
		if !quiet {
			if err := output.WriteItems(cmd.OutOrStdout(), []domain.Item{newItem}, trayNames, strings.TrimSpace(sess.UserID), by, format); err != nil {
				return err
			}
		}
		if hookCfg != nil {
			srcName := strings.TrimSpace(by[strings.TrimSpace(newItem.SourceUserID)])
			for i, r := range hookCfg.Hooks {
				if strings.TrimSpace(r.Event) != listenhook.EventItemPending {
					continue
				}
				tf := ""
				if i < len(ruleTray) {
					tf = ruleTray[i]
				}
				if !listenhook.MatchPending(r, newItem, owned, sess.UserID, tf) {
					continue
				}
				runHookWithLogs(cmd, r, listenhook.EventItemPending, sess, newItem, srcName)
			}
		}
	}

	if needOutbox && hasOld && hasNew && strings.EqualFold(ch.Type, "UPDATE") {
		trans := listenhook.OutboxTransitions{}
		oldStatus := strings.TrimSpace(oldItem.Status)
		newStatus := strings.TrimSpace(newItem.Status)
		if !strings.EqualFold(oldStatus, newStatus) {
			switch strings.ToLower(newStatus) {
			case "completed":
				trans.Completed = []domain.Item{newItem}
			case "accepted":
				trans.Accepted = []domain.Item{newItem}
			case "declined":
				trans.Declined = []domain.Item{newItem}
			}
		}
		if len(trans.Completed)+len(trans.Accepted)+len(trans.Declined) > 0 && hookCfg != nil {
			runOutboxHooks(cmd, svcs, sess, hookCfg, ruleTray, trans)
		}
	}
	return nil
}

func emitListenEvent(cmd *cobra.Command, event string, it domain.Item, trayNames map[string]string, currentUserID string, displayByID map[string]string) {
	tn := strings.TrimSpace(trayNames[strings.TrimSpace(it.TrayID)])
	if tn == "" {
		tn = strings.TrimSpace(it.TrayID)
	}
	by := output.FormatSourceUser(it.SourceUserID, currentUserID, displayByID)
	title := strconv.Quote(strings.TrimSpace(it.Title))
	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "event %s tray=%s by=%s title=%s\n", strings.TrimSpace(event), tn, by, title)
}

func runHookWithLogs(cmd *cobra.Command, r listenhook.Rule, event string, sess domain.Session, it domain.Item, srcName string) {
	cmdLabel := strings.Join(r.Command, " ")
	if cmdLabel == "" {
		cmdLabel = "<empty command>"
	}
	verbose, _ := cmd.Flags().GetBool("verbose")
	if verbose {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "hook matched event=%s command=%q item=%s\n", strings.TrimSpace(event), cmdLabel, strings.TrimSpace(it.ID))
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "hook run event=%s command=%q\n", strings.TrimSpace(event), cmdLabel)
	}
	if err := listenhook.RunHook(r, event, sess, it, srcName); err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "hook failed event=%s command=%q err=%v\n", strings.TrimSpace(event), cmdLabel, err)
		return
	}
	if verbose {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "hook ok event=%s command=%q\n", strings.TrimSpace(event), cmdLabel)
	}
}

func itemFromRealtime(m map[string]any) (domain.Item, bool) {
	if len(m) == 0 {
		return domain.Item{}, false
	}
	var it domain.Item
	it.ID, _ = asString(m["id"])
	it.TrayID, _ = asString(m["tray_id"])
	it.SourceUserID, _ = asString(m["source_user_id"])
	it.Title, _ = asString(m["title"])
	it.Status, _ = asString(m["status"])
	if v, ok := asString(m["decline_reason"]); ok && strings.TrimSpace(v) != "" {
		it.DeclineReason = &v
	}
	if t, ok := asTime(m["accepted_at"]); ok {
		it.AcceptedAt = &t
	}
	if t, ok := asTime(m["declined_at"]); ok {
		it.DeclinedAt = &t
	}
	if t, ok := asTime(m["completed_at"]); ok {
		it.CompletedAt = &t
	}
	if v, ok := asInt(m["sort_order"]); ok {
		it.SortOrder = v
	}
	if strings.TrimSpace(it.ID) == "" {
		return domain.Item{}, false
	}
	return it, true
}

func asInt(v any) (int, bool) {
	switch t := v.(type) {
	case int:
		return t, true
	case int64:
		return int(t), true
	case float64:
		return int(t), true
	default:
		return 0, false
	}
}

func asString(v any) (string, bool) {
	if v == nil {
		return "", false
	}
	switch t := v.(type) {
	case string:
		return t, true
	default:
		return fmt.Sprintf("%v", v), true
	}
}

func asTime(v any) (time.Time, bool) {
	s, ok := asString(v)
	if !ok || strings.TrimSpace(s) == "" {
		return time.Time{}, false
	}
	t, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(s))
	if err != nil {
		t, err = time.Parse(time.RFC3339, strings.TrimSpace(s))
		if err != nil {
			return time.Time{}, false
		}
	}
	return t, true
}
