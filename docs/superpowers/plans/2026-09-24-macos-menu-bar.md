# macOS `tray bar` Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship a macOS-only `tray bar` command that shows a menu-bar badge for local open + remote pending items, lists them grouped by tray, copies `tray-ref item-id` on click, and document the accept→work→complete handoff in the existing `tray-cli` skill.

**Architecture:** Pure aggregator in `internal/bar` builds snapshots (badge, sections, clipboard strings). Thin loaders pull `localtray.Store.AllOpenItems()` and owned remote pending via existing query helpers. Darwin UI uses `fyne.io/systray` (CGO) behind build tags; other OS get a friendly error. Optional `--daemon` mirrors `listen` pid/log. Skill + user docs describe the clipboard handoff—no new skill.

**Tech Stack:** Go 1.25, cobra, `fyne.io/systray`, `github.com/atotto/clipboard` (already indirect), existing `localtray` + PostgREST item APIs.

**Spec:** [`docs/superpowers/specs/2026-09-24-macos-menu-bar-design.md`](../specs/2026-09-24-macos-menu-bar-design.md)

---

## File map

| Path | Role |
|------|------|
| `internal/bar/snapshot.go` | Types + pure `BuildSnapshot` / `ClipboardLine` |
| `internal/bar/snapshot_test.go` | Unit tests for grouping, badge, clipboard, collisions |
| `internal/bar/local_ref.go` | Pure `LocalTrayRef(rec localtray.TrayRecord) string` |
| `internal/bar/load.go` | I/O: load local rows + remote pending into `[]Item` inputs (thin) |
| `internal/cli/commands/bar.go` | `cmdBar` / `runBar` flags, refresh loop, wire loaders → UI |
| `internal/cli/commands/bar_daemon.go` | `bar.pid` / `bar.log` (mirror listen) |
| `internal/cli/commands/bar_ui_darwin.go` | `//go:build darwin` systray runner |
| `internal/cli/commands/bar_ui_stub.go` | `//go:build !darwin` error stub |
| `internal/cli/commands/register.go` | Register `cmdBar()` |
| `scripts/build-release.sh` | Darwin builds with `CGO_ENABLED=1` on macOS hosts |
| `docs/user/menu-bar.md` | End-user docs |
| `docs/user/README.md`, `README.md` | Links |
| `docs/maintainers/distribution.md` | CGO / darwin build note |
| `skills/tray-cli/SKILL.md` | Menu bar + accept-first resolve flow |

---

### Task 1: Pure snapshot types + `BuildSnapshot`

**Files:**
- Create: `internal/bar/snapshot.go`
- Create: `internal/bar/snapshot_test.go`

- [ ] **Step 1: Write the failing test**

```go
package bar_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tjdsneto/tray-cli/internal/bar"
)

func TestBuildSnapshot_groupsByTray_badgeAndNewestFirst(t *testing.T) {
	t.Parallel()
	older := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	newer := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	in := []bar.Item{
		{Scope: bar.ScopeLocal, TrayRef: "inbox", TrayName: "inbox", ItemID: "loc1", Title: "old", CreatedAt: older},
		{Scope: bar.ScopeLocal, TrayRef: "inbox", TrayName: "inbox", ItemID: "loc2", Title: "new", CreatedAt: newer},
		{Scope: bar.ScopeRemote, TrayRef: "work", TrayName: "work", ItemID: "rem1", Title: "pending", CreatedAt: newer},
	}
	snap := bar.BuildSnapshot(in, "")
	require.Equal(t, 3, snap.Badge)
	require.Len(t, snap.Sections, 2)
	require.Equal(t, "inbox", snap.Sections[0].Header)
	require.Equal(t, "loc2", snap.Sections[0].Rows[0].ItemID)
	require.Equal(t, "work", snap.Sections[1].Header)
}

func TestBuildSnapshot_nameCollision_disambiguates(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	in := []bar.Item{
		{Scope: bar.ScopeLocal, TrayRef: "inbox", TrayName: "inbox", ItemID: "l1", Title: "a", CreatedAt: now},
		{Scope: bar.ScopeRemote, TrayRef: "inbox", TrayName: "inbox", ItemID: "r1", Title: "b", CreatedAt: now},
	}
	snap := bar.BuildSnapshot(in, "")
	require.Equal(t, 2, snap.Badge)
	require.Len(t, snap.Sections, 2)
	headers := []string{snap.Sections[0].Header, snap.Sections[1].Header}
	require.Contains(t, headers, "inbox · local")
	require.Contains(t, headers, "inbox · remote")
}

func TestClipboardLine(t *testing.T) {
	t.Parallel()
	require.Equal(t, "inbox abc123", bar.ClipboardLine(bar.Row{TrayRef: "inbox", ItemID: "abc123"}))
}

func TestBuildSnapshot_remoteErrLine(t *testing.T) {
	t.Parallel()
	snap := bar.BuildSnapshot(nil, "remote unavailable")
	require.Equal(t, 0, snap.Badge)
	require.Equal(t, "remote unavailable", snap.RemoteStatus)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/bar/ -count=1`

Expected: FAIL — package `bar` does not exist / undefined symbols.

- [ ] **Step 3: Write minimal implementation**

Create `internal/bar/snapshot.go`:

```go
package bar

import (
	"fmt"
	"sort"
	"time"
)

type Scope string

const (
	ScopeLocal  Scope = "local"
	ScopeRemote Scope = "remote"
)

// Item is one attention-backlog row before grouping.
type Item struct {
	Scope     Scope
	TrayRef   string // first clipboard token
	TrayName  string // group key (display name before disambiguation)
	ItemID    string
	Title     string
	CreatedAt time.Time
}

type Row struct {
	Scope   Scope
	TrayRef string
	ItemID  string
	Title   string
}

type Section struct {
	Header string
	Rows   []Row
}

type Snapshot struct {
	Badge        int
	Sections     []Section
	RemoteStatus string // muted status line; empty if OK
}

func ClipboardLine(r Row) string {
	return fmt.Sprintf("%s %s", r.TrayRef, r.ItemID)
}

// BuildSnapshot groups items by tray name (A–Z), newest-first within a section.
// When the same TrayName appears in both scopes, headers become "name · local" / "name · remote".
func BuildSnapshot(items []Item, remoteStatus string) Snapshot {
	type key struct {
		name  string
		scope Scope
	}
	nameScopes := map[string]map[Scope]struct{}{}
	for _, it := range items {
		if nameScopes[it.TrayName] == nil {
			nameScopes[it.TrayName] = map[Scope]struct{}{}
		}
		nameScopes[it.TrayName][it.Scope] = struct{}{}
	}
	buckets := map[key][]Item{}
	for _, it := range items {
		k := key{name: it.TrayName, scope: it.Scope}
		buckets[k] = append(buckets[k], it)
	}
	var keys []key
	for k := range buckets {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].name != keys[j].name {
			return keys[i].name < keys[j].name
		}
		return keys[i].scope < keys[j].scope
	})
	snap := Snapshot{Badge: len(items), RemoteStatus: remoteStatus}
	for _, k := range keys {
		list := buckets[k]
		sort.Slice(list, func(i, j int) bool {
			return list[i].CreatedAt.After(list[j].CreatedAt)
		})
		header := k.name
		if len(nameScopes[k.name]) > 1 {
			header = fmt.Sprintf("%s · %s", k.name, k.scope)
		}
		sec := Section{Header: header}
		for _, it := range list {
			sec.Rows = append(sec.Rows, Row{
				Scope: it.Scope, TrayRef: it.TrayRef, ItemID: it.ItemID, Title: it.Title,
			})
		}
		snap.Sections = append(snap.Sections, sec)
	}
	return snap
}
```

- [ ] **Step 4: Run tests and make sure they pass**

Run: `go test ./internal/bar/ -count=1`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/bar/snapshot.go internal/bar/snapshot_test.go
git commit -m "$(cat <<'EOF'
feat(bar): pure snapshot builder for menu bar backlog

EOF
)"
```

---

### Task 2: Local clipboard tray-ref helper

**Files:**
- Create: `internal/bar/local_ref.go`
- Create: `internal/bar/local_ref_test.go`

- [ ] **Step 1: Write the failing test**

```go
package bar_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tjdsneto/tray-cli/internal/bar"
	"github.com/tjdsneto/tray-cli/internal/localtray"
)

func TestLocalTrayRef(t *testing.T) {
	t.Parallel()
	require.Equal(t, "inbox", bar.LocalTrayRef(localtray.TrayRecord{
		Kind: localtray.KindGlobal, Name: "inbox", ID: "global:inbox",
	}))
	require.Equal(t, "dir:/Users/me/proj", bar.LocalTrayRef(localtray.TrayRecord{
		Kind: localtray.KindDir, Path: "/Users/me/proj", ID: "dir:/Users/me/proj",
	}))
	require.Equal(t, "branch:abc:feat", bar.LocalTrayRef(localtray.TrayRecord{
		Kind: localtray.KindBranch, ID: "branch:abc:feat",
	}))
}
```

(Adjust expected dir/branch strings to match whatever `LocalTrayRef` documents—prefer values agents can paste into CLI / resolve.)

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/bar/ -run TestLocalTrayRef -count=1`

Expected: FAIL — `LocalTrayRef` undefined.

- [ ] **Step 3: Write minimal implementation**

```go
package bar

import (
	"strings"

	"github.com/tjdsneto/tray-cli/internal/localtray"
)

// LocalTrayRef returns the first clipboard token for a local tray.
// Global trays use the bare name; dir/branch use the canonical tray id (unambiguous for agents).
func LocalTrayRef(rec localtray.TrayRecord) string {
	switch rec.Kind {
	case localtray.KindGlobal:
		return strings.TrimSpace(rec.Name)
	default:
		return strings.TrimSpace(rec.ID)
	}
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./internal/bar/ -count=1`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/bar/local_ref.go internal/bar/local_ref_test.go
git commit -m "$(cat <<'EOF'
feat(bar): clipboard tray-ref for local trays

EOF
)"
```

---

### Task 3: Load local + remote items into `[]bar.Item`

**Files:**
- Create: `internal/bar/load.go`
- Create: `internal/bar/load_test.go` (table test converting fake structs—or keep load thin and test via small helpers)

Prefer extracting pure converters so I/O stays thin:

```go
// ItemsFromLocal maps store rows to bar.Item (pure).
func ItemsFromLocal(rows []localtray.ItemWithTray) []Item

// ItemsFromRemote maps domain items + name-by-tray-id to bar.Item (pure).
func ItemsFromRemote(items []domain.Item, trayNames map[string]string) []Item
```

- [ ] **Step 1: Write failing tests for converters**

```go
func TestItemsFromLocal(t *testing.T) {
	t.Parallel()
	rows := []localtray.ItemWithTray{{
		Item: localtray.Item{ID: "aabb", Title: "t", CreatedAt: time.Unix(1, 0).UTC()},
		Tray: localtray.TrayRecord{Kind: localtray.KindGlobal, Name: "inbox", ID: "global:inbox"},
	}}
	out := bar.ItemsFromLocal(rows)
	require.Len(t, out, 1)
	require.Equal(t, bar.ScopeLocal, out[0].Scope)
	require.Equal(t, "inbox", out[0].TrayRef)
	require.Equal(t, "inbox", out[0].TrayName)
}

func TestItemsFromRemote(t *testing.T) {
	t.Parallel()
	items := []domain.Item{{
		ID: "uuid-1", TrayID: "tray-uuid", Title: "p", CreatedAt: time.Unix(2, 0).UTC(),
	}}
	names := map[string]string{"tray-uuid": "work"}
	out := bar.ItemsFromRemote(items, names)
	require.Equal(t, "work", out[0].TrayRef)
	require.Equal(t, bar.ScopeRemote, out[0].Scope)
}
```

- [ ] **Step 2: Run — expect FAIL**

Run: `go test ./internal/bar/ -run 'ItemsFrom' -count=1`

- [ ] **Step 3: Implement converters in `load.go`**

```go
package bar

import (
	"strings"

	"github.com/tjdsneto/tray-cli/internal/domain"
	"github.com/tjdsneto/tray-cli/internal/localtray"
)

func ItemsFromLocal(rows []localtray.ItemWithTray) []Item {
	out := make([]Item, 0, len(rows))
	for _, r := range rows {
		name := strings.TrimSpace(r.Tray.Name)
		if name == "" {
			name = localtray.DisplayLabel(r.Tray.ID)
		}
		out = append(out, Item{
			Scope: ScopeLocal, TrayRef: LocalTrayRef(r.Tray), TrayName: name,
			ItemID: r.Item.ID, Title: r.Item.Title, CreatedAt: r.Item.CreatedAt,
		})
	}
	return out
}

func ItemsFromRemote(items []domain.Item, trayNames map[string]string) []Item {
	out := make([]Item, 0, len(items))
	for _, it := range items {
		name := strings.TrimSpace(trayNames[it.TrayID])
		if name == "" {
			name = it.TrayID
		}
		out = append(out, Item{
			Scope: ScopeRemote, TrayRef: name, TrayName: name,
			ItemID: it.ID, Title: it.Title, CreatedAt: it.CreatedAt,
		})
	}
	return out
}
```

Also add `LoadSnapshot` in the same file (or `load_io.go`) used by the command—**not** unit-tested heavily:

```go
// LoadAttentionItems returns local open + remote pending (if sess non-nil).
// On remote error, returns local items and a remoteStatus message (never a hard error for remote).
func LoadAttentionItems(ctx context.Context, configDir string, svcs *domain.Services, sess *domain.Session, aliases map[string]string) (items []Item, remoteStatus string, err error)
```

Implementation sketch:
1. `store := localtray.NewStore(configDir)`; `AllOpenItems()` → `ItemsFromLocal` (local failure **is** a hard error).
2. If `sess == nil` or `svcs == nil`, return local only.
3. Else build query with logic equivalent to `pendingItemsOnOwnedTraysQuery` — **prefer calling a small exported helper** from `commands` only if import cycles allow; otherwise duplicate the query construction inside `bar` by accepting `[]domain.Item` from the CLI layer.

**Import-cycle rule:** Keep `internal/bar` free of `internal/cli/commands`. Put the remote fetch in `commands/bar.go` (or `commands/bar_load.go`) and pass `[]domain.Item` + name map into `ItemsFromRemote` + `BuildSnapshot`.

- [ ] **Step 4: Run `go test ./internal/bar/ -count=1`** — PASS

- [ ] **Step 5: Commit**

```bash
git add internal/bar/
git commit -m "$(cat <<'EOF'
feat(bar): convert local and remote items for snapshots

EOF
)"
```

---

### Task 4: `tray bar` command stub (all platforms) + register

**Files:**
- Create: `internal/cli/commands/bar.go`
- Create: `internal/cli/commands/bar_ui_stub.go` (`//go:build !darwin`)
- Create: `internal/cli/commands/bar_ui_darwin.go` (`//go:build darwin`) — temporary stub that returns `"not implemented"` until Task 5
- Modify: `internal/cli/commands/register.go` — add `root.AddCommand(cmdBar())` after listen section
- Test: `internal/cli/commands/bar_test.go` — Use string / flags only

- [ ] **Step 1: Write failing command test**

```go
func TestCmdBar_use(t *testing.T) {
	c := cmdBar()
	require.Equal(t, "bar", c.Name())
	require.NotNil(t, c.Flags().Lookup("interval"))
	require.NotNil(t, c.Flags().Lookup("daemon"))
}
```

- [ ] **Step 2: Run — FAIL** (undefined)

- [ ] **Step 3: Implement `cmdBar` + stubs**

`bar.go`:

```go
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
```

`bar_ui_stub.go`:

```go
//go:build !darwin

func runBarUI(cmd *cobra.Command) error {
	return fmt.Errorf("tray bar is only supported on macOS")
}
```

Darwin temporary:

```go
//go:build darwin

func runBarUI(cmd *cobra.Command) error {
	return fmt.Errorf("tray bar UI not wired yet")
}
```

Register after `cmdListen()`.

- [ ] **Step 4: `go test ./internal/cli/commands/ -run TestCmdBar -count=1`** — PASS

- [ ] **Step 5: Commit**

```bash
git add internal/cli/commands/bar.go internal/cli/commands/bar_ui_*.go internal/cli/commands/bar_test.go internal/cli/commands/register.go
git commit -m "$(cat <<'EOF'
feat(cli): add tray bar command stub (macOS-only)

EOF
)"
```

---

### Task 5: Wire refresh + snapshot into darwin UI (systray)

**Files:**
- Modify: `internal/cli/commands/bar.go` — load loop, auth optional
- Modify: `internal/cli/commands/bar_ui_darwin.go` — full systray
- Run: `go get fyne.io/systray@latest` and `go get github.com/atotto/clipboard@latest` (promote direct)

- [ ] **Step 1: Add dependencies**

```bash
go get fyne.io/systray@latest
go get github.com/atotto/clipboard@latest
go mod tidy
```

- [ ] **Step 2: Implement refresh + apply in `bar.go`**

Shared logic (no build tag):

```go
type barRuntime struct {
	interval time.Duration
	configDir string
	// last atomic.Pointer[bar.Snapshot]
}

func (r *barRuntime) refresh(cmd *cobra.Command) bar.Snapshot {
	ctx := cmd.Context()
	localRows, err := localtray.NewStore(r.configDir).AllOpenItems()
	if err != nil {
		// log to cmd.ErrOrStderr(); keep previous snapshot if any
		return r.lastOrEmpty()
	}
	items := bar.ItemsFromLocal(localRows)
	remoteStatus := ""
	sess, authErr := /* try optional auth without failing */; 
	// Pattern: if credentials present, RequireAuth / EnsureFresh; on failure set remoteStatus = "remote unavailable"
	if authErr == nil && sess != nil {
		svcs := /* build services like other commands */
		q, qerr := pendingItemsOnOwnedTraysQuery(ctx, svcs, *sess, "", aliases)
		if qerr != nil {
			remoteStatus = "remote unavailable"
		} else {
			list, lerr := svcs.Items.List(ctx, *sess, q)
			// prefer OrderCreated: "desc" — set on q before List if not already
			if lerr != nil {
				remoteStatus = "remote unavailable"
			} else {
				owned, _ := svcs.Trays.ListOwned(ctx, *sess)
				names := trayref.TrayNameMap(owned)
				items = append(items, bar.ItemsFromRemote(list, names)...)
			}
		}
	}
	snap := bar.BuildSnapshot(items, remoteStatus)
	r.store(snap)
	return snap
}
```

Match existing command patterns for building `domain.Services` and session (copy from `listen.go` / `review.go`; keep optional auth so signed-out local works).

For newest-first remote order, set `q.OrderCreated = "desc"` when listing for the bar (or sort in `BuildSnapshot`—already sorts by `CreatedAt`).

- [ ] **Step 3: Implement systray UI on darwin**

```go
//go:build darwin

package commands

import (
	"fmt"

	"fyne.io/systray"
	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
	"github.com/tjdsneto/tray-cli/internal/bar"
)

func runBarUI(cmd *cobra.Command) error {
	interval, _ := cmd.Flags().GetDuration("interval")
	daemon, _ := cmd.Flags().GetBool("daemon")
	configDir, err := cmdDeps.ConfigDir()
	if err != nil {
		return err
	}
	if daemon {
		// Task 6 wires this; for now call helpers if present
	}
	rt := &barRuntime{interval: interval, configDir: configDir}

	systray.Run(func() {
		systray.SetTitle("tray")
		systray.SetTooltip("tray")
		// optional: systray.SetTemplateIcon(icon, icon) with embedded template PNG
		mRemote := systray.AddMenuItem("", "")
		mRemote.Hide()
		var itemMenus []*systray.MenuItem
		systray.AddSeparator()
		mRefresh := systray.AddMenuItem("Refresh", "")
		mQuit := systray.AddMenuItem("Quit", "")

		apply := func(snap bar.Snapshot) {
			if snap.Badge > 0 {
				systray.SetTitle(fmt.Sprintf("tray %d", snap.Badge))
			} else {
				systray.SetTitle("tray")
			}
			// Rebuild: simplest v1 — Quit+Relaunch not needed; remove old itemMenus and re-add
			// (fyne systray supports Hide on old items; or rebuild on each refresh by hiding extras)
			_ = mRemote
			_ = itemMenus
			_ = snap
			// Concrete approach: keep a pool of MenuItems; Hide unused; SetTitle on each row;
			// on ClickedCh, clipboard.WriteAll(bar.ClipboardLine(row))
		}

		apply(rt.refresh(cmd))
		go func() {
			t := time.NewTicker(rt.interval)
			defer t.Stop()
			for {
				select {
				case <-cmd.Context().Done():
					systray.Quit()
					return
				case <-t.C:
					apply(rt.refresh(cmd))
				case <-mRefresh.ClickedCh:
					apply(rt.refresh(cmd))
				case <-mQuit.ClickedCh:
					systray.Quit()
					return
				}
			}
		}()
	}, func() {})
	return nil
}
```

**Menu rebuild note:** Implement a fixed max pool (e.g. 50 row slots + section headers as disabled items) created once; on apply, set titles/visibility. Click handlers close over current `bar.Row` via mutex-protected slice.

- [ ] **Step 4: Manual smoke on macOS**

```bash
go run -tags '' ./cmd/tray -- bar
# add local item: tray add "test from bar" inbox
# confirm badge and menu; click copies "inbox <id>"; pbpaste
```

Expected: Icon titled `tray N`; click copies two tokens.

- [ ] **Step 5: Commit**

```bash
git add go.mod go.sum internal/cli/commands/bar.go internal/cli/commands/bar_ui_darwin.go
git commit -m "$(cat <<'EOF'
feat(bar): macOS systray UI with refresh and clipboard copy

EOF
)"
```

---

### Task 6: Daemon mode (`bar.pid` / `bar.log`)

**Files:**
- Create: `internal/cli/commands/bar_daemon.go` (reuse patterns from `listen_daemon.go`; can share `pidAlive` already in package)
- Modify: `bar.go` / `bar_ui_darwin.go` to call acquire + redirect when `--daemon`

- [ ] **Step 1: Implement**

```go
const (
	barPidFile = "bar.pid"
	barLogFile = "bar.log"
)

func acquireBarDaemon(configDir string) (cleanup func(), err error) { /* same as listen, different filename */ }
func redirectBarDaemonLog(cmd *cobra.Command, configDir string) (cleanup func(), err error) { /* same */ }
```

In `runBar` before UI:

```go
if daemon {
	cleanup, err := acquireBarDaemon(configDir)
	if err != nil { return err }
	defer cleanup()
	logCleanup, err := redirectBarDaemonLog(cmd, configDir)
	if err != nil { return err }
	defer logCleanup()
}
```

- [ ] **Step 2: Manual check**

```bash
tray bar --daemon
test -f "$TRAY_CONFIG_DIR/bar.pid"
# second start should error "already running"
```

- [ ] **Step 3: Commit**

```bash
git add internal/cli/commands/bar_daemon.go internal/cli/commands/bar.go
git commit -m "$(cat <<'EOF'
feat(bar): daemon pid and log files

EOF
)"
```

---

### Task 7: Release builds + maintainer docs (CGO on darwin)

**Files:**
- Modify: `scripts/build-release.sh`
- Modify: `docs/maintainers/distribution.md`

- [ ] **Step 1: Update `build_one` for darwin CGO**

```bash
build_one() {
	local goos="$1"
	local goarch="$2"
	local cgo=0
	if [[ "${goos}" == "darwin" ]]; then
		if [[ "$(uname -s)" != "Darwin" ]]; then
			echo "Skipping ${goos}/${goarch}: tray bar needs CGO; build darwin artifacts on macOS."
			return 0
		fi
		cgo=1
	fi
	# ...
	GOOS="${goos}" GOARCH="${goarch}" CGO_ENABLED="${cgo}" go build -trimpath \
		-ldflags "${LDFLAGS}" \
		-o "${tmp}/tray" ./cmd/tray
}
```

Note: native `GOARCH` must match when CGO=1 (or use `arch -arm64` / dual CI jobs). Document: publish darwin tarballs from a Mac runner; linux stays `CGO_ENABLED=0`.

- [ ] **Step 2: Document in `docs/maintainers/distribution.md`**

Add a short subsection: darwin release binaries enable CGO for `tray bar` / systray; must be compiled on macOS; linux remains static.

- [ ] **Step 3: Commit**

```bash
git add scripts/build-release.sh docs/maintainers/distribution.md
git commit -m "$(cat <<'EOF'
build: enable CGO for darwin release artifacts (tray bar)

EOF
)"
```

---

### Task 8: User docs + `tray-cli` skill

**Files:**
- Create: `docs/user/menu-bar.md`
- Modify: `docs/user/README.md`, `README.md` (one link each)
- Modify: `skills/tray-cli/SKILL.md`

- [ ] **Step 1: Write `docs/user/menu-bar.md`**

Cover: `tray bar` / `--daemon` / `--interval`; badge = local open + remote pending; click copies `tray-ref item-id`; accept drops badge; macOS only; handoff to agents.

- [ ] **Step 2: Link from `docs/user/README.md` and root `README.md`**

- [ ] **Step 3: Extend `skills/tray-cli/SKILL.md`**

Add to command map Automation row: `` `bar` (macOS menu bar) ``.

Add section **Menu bar handoff**:

- User may paste `tray-ref item-id` from the menu bar.
- Fetch details with CLI; if remote **pending**, run **`tray accept <id>`** before work; then **`tray complete <id>`**.
- Local open: skip accept.
- Do not invent status.

- [ ] **Step 4: Commit**

```bash
git add docs/user/menu-bar.md docs/user/README.md README.md skills/tray-cli/SKILL.md
git commit -m "$(cat <<'EOF'
docs: document tray bar and skill handoff (accept then complete)

EOF
)"
```

---

### Task 9: Final verification

- [ ] **Step 1: Unit tests**

Run: `go test ./internal/bar/ ./internal/cli/commands/ -count=1`

Expected: PASS (darwin UI not required in unit tests).

- [ ] **Step 2: Build on macOS**

Run: `CGO_ENABLED=1 go build -o /tmp/tray ./cmd/tray && /tmp/tray bar --help`

Expected: help text; no crash.

- [ ] **Step 3: Spec checklist (manual)**

| Spec item | Verify |
|-----------|--------|
| Badge local open + remote pending | Add both; see count |
| Group by tray; collision labels | Same name local+remote |
| Click → `tray id` clipboard | `pbpaste` |
| Accept decrements badge | `tray accept`; wait interval / Refresh |
| Accepted leaves menu | Row gone |
| Skill documents flow | Read SKILL.md |
| Non-darwin message | Cross-compile stub or read `bar_ui_stub.go` |

- [ ] **Step 4: Commit any fixes** from verification, or stop if clean.

---

## Spec coverage (self-review)

| Spec requirement | Task |
|------------------|------|
| Badge local open + remote pending | 1, 3, 5 |
| Menu grouped by tray; newest-first; collision labels | 1 |
| Accepted excluded (only pending/open loaded) | 3, 5 |
| Clipboard `tray-ref item-id` | 1, 2, 5 |
| `tray bar` + interval + daemon | 4, 5, 6 |
| macOS only | 4 stub |
| Separate from listen | 4–6 |
| Extend `tray-cli` skill (no new skill) | 8 |
| User + distribution docs | 7, 8 |
| Pure aggregator unit tests | 1–3 |
| CGO darwin releases | 7 |
| Non-goals (full UI, in-menu triage, LaunchAgent) | Not planned |

**Ambiguity resolved in plan:** Local dir/branch clipboard refs use canonical tray **id**; global uses bare **name**.

**Placeholder scan:** No TBD steps; systray menu pool described concretely in Task 5.
