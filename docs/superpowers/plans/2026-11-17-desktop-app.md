# Tray desktop app Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship a macOS desktop app that triages trays in a three-pane window and owns a menu bar with today's copy-handoff behavior, both reading one snapshot from a loopback `tray api` process.

**Architecture:** Pure rules and JSON shapes live in `internal/api`. A thin `net/http` handler owns the snapshot and calls a `Source`. `tray api` is a cobra command that starts that handler on `127.0.0.1`, prints one JSON handshake line, and exits when stdin closes. The Tauri app in `desktop/` spawns `tray` from `PATH`, then renders the window and the status item from that API. Phase 1 is usable with `curl` before the window exists.

**Tech Stack:** Go 1.25, cobra, existing `internal/bar`, `internal/localtray`, and domain services. Tauri 2, React 18, Vite, Tailwind, Vitest.

**Spec:** [`docs/superpowers/specs/2026-11-17-desktop-app-design.md`](../specs/2026-11-17-desktop-app-design.md)

---

## File map

| Path | Role |
|------|------|
| `internal/api/actions.go` | Which actions a scope and status allow |
| `internal/api/snooze.go` | Preset → RFC3339 instant |
| `internal/api/dto.go` | Snapshot, tray, item, and error JSON |
| `internal/api/filter.go` | Open-work filter and newest-first sort |
| `internal/api/handler.go` | Bearer auth, routes, snapshot rebuild |
| `internal/api/source.go` | `Source` interface the handler calls |
| `internal/cli/commands/api.go` | `tray api` process lifecycle |
| `internal/cli/commands/api_source.go` | `Source` over local files and domain services |
| `internal/cli/commands/register.go` | Register `cmdAPI()` |
| `desktop/` | Tauri window, menu bar, Vitest for the action buttons |
| `docs/user/desktop.md` | How to launch the app |

Do not implement invite, members, listen, prune, upgrade, reorder, remove, a sign-in screen, a settings screen, or body editing.

---

### Task 1: Action table

**Files:**
- Create: `internal/api/actions.go`
- Test: `internal/api/actions_test.go`

- [ ] **Step 1: Write the failing test**

```go
package api_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tjdsneto/tray-cli/internal/api"
)

func TestActionsFor(t *testing.T) {
	t.Parallel()
	require.Equal(t, []string{"complete"}, api.ActionsFor("local", "open"))
	require.Equal(t, []string{"accept", "decline", "snooze"}, api.ActionsFor("remote", "pending"))
	require.Equal(t, []string{"complete", "archive"}, api.ActionsFor("remote", "accepted"))
	require.Equal(t, []string{"archive"}, api.ActionsFor("remote", "snoozed"))
	require.Empty(t, api.ActionsFor("local", "done"))
	require.Empty(t, api.ActionsFor("remote", "declined"))
	require.Empty(t, api.ActionsFor("local", "pending"))
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/api/ -count=1 -race -run TestActionsFor`
Expected: FAIL with `undefined: api.ActionsFor` or `no Go files`

- [ ] **Step 3: Write minimal implementation**

```go
package api

func ActionsFor(scope, status string) []string {
	switch {
	case scope == "local" && status == "open":
		return []string{"complete"}
	case scope == "remote" && status == "pending":
		return []string{"accept", "decline", "snooze"}
	case scope == "remote" && status == "accepted":
		return []string{"complete", "archive"}
	case scope == "remote" && status == "snoozed":
		return []string{"archive"}
	default:
		return []string{}
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/api/ -count=1 -race -run TestActionsFor`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/api/actions.go internal/api/actions_test.go
git commit -m "feat: define which desktop actions a tray item allows"
```

---

### Task 2: Snooze presets

**Files:**
- Create: `internal/api/snooze.go`
- Test: `internal/api/snooze_test.go`

- [ ] **Step 1: Write the failing test**

```go
package api_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tjdsneto/tray-cli/internal/api"
)

func TestSnoozeUntil(t *testing.T) {
	t.Parallel()
	loc := time.FixedZone("UTC-3", -3*60*60)
	now := time.Date(2026, 6, 15, 18, 30, 0, 0, loc)

	hour, err := api.SnoozeUntil("hour", now)
	require.NoError(t, err)
	require.True(t, hour.Equal(now.Add(time.Hour)))

	tomorrow, err := api.SnoozeUntil("tomorrow", now)
	require.NoError(t, err)
	require.True(t, tomorrow.Equal(time.Date(2026, 6, 16, 9, 0, 0, 0, loc)))

	week, err := api.SnoozeUntil("week", now)
	require.NoError(t, err)
	require.True(t, week.Equal(time.Date(2026, 6, 22, 9, 0, 0, 0, loc)))

	_, err = api.SnoozeUntil("calendar", now)
	require.Error(t, err)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/api/ -count=1 -race -run TestSnoozeUntil`
Expected: FAIL with `undefined: api.SnoozeUntil`

- [ ] **Step 3: Write minimal implementation**

```go
package api

import (
	"fmt"
	"time"
)

func SnoozeUntil(preset string, now time.Time) (time.Time, error) {
	loc := now.Location()
	y, m, d := now.Date()
	switch preset {
	case "hour":
		return now.Add(time.Hour), nil
	case "tomorrow":
		return time.Date(y, m, d, 9, 0, 0, 0, loc).AddDate(0, 0, 1), nil
	case "week":
		return time.Date(y, m, d, 9, 0, 0, 0, loc).AddDate(0, 0, 7), nil
	default:
		return time.Time{}, fmt.Errorf("unknown snooze preset %q", preset)
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/api/ -count=1 -race -run TestSnoozeUntil`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/api/snooze.go internal/api/snooze_test.go
git commit -m "feat: map desktop snooze presets to an instant"
```

---

### Task 3: JSON shapes, open-work filter, newest-first

**Files:**
- Create: `internal/api/dto.go`
- Create: `internal/api/filter.go`
- Test: `internal/api/dto_test.go`

- [ ] **Step 1: Write the failing test**

```go
package api_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tjdsneto/tray-cli/internal/api"
	"github.com/tjdsneto/tray-cli/internal/bar"
)

func TestSnapshotFrom_includesClipboard(t *testing.T) {
	t.Parallel()
	in := bar.BuildSnapshot([]bar.Item{{
		Scope: bar.ScopeLocal, TrayRef: "notes", TrayName: "notes",
		ItemID: "abc", Title: "Draft", CreatedAt: time.Unix(1, 0).UTC(),
	}}, "")
	out := api.SnapshotFrom(in)
	require.Equal(t, 1, out.Badge)
	require.Equal(t, "notes", out.Sections[0].Header)
	require.Equal(t, "notes abc", out.Sections[0].Rows[0].Clipboard)
	require.Equal(t, "abc", out.Sections[0].Rows[0].ItemID)
}

func TestFilterOpenAndSortNewest(t *testing.T) {
	t.Parallel()
	older := time.Unix(1, 0).UTC()
	newer := time.Unix(2, 0).UTC()
	items := []api.Item{
		{ID: "a", Scope: "local", Status: "done", CreatedAt: newer},
		{ID: "b", Scope: "local", Status: "open", CreatedAt: older},
		{ID: "c", Scope: "remote", Status: "pending", CreatedAt: newer},
		{ID: "d", Scope: "remote", Status: "completed", CreatedAt: newer},
		{ID: "e", Scope: "remote", Status: "snoozed", CreatedAt: older},
	}
	open := api.SelectItems(items, "open")
	require.Equal(t, []string{"c", "b", "e"}, ids(open))
	all := api.SelectItems(items, "all")
	require.Len(t, all, 5)
	require.Equal(t, "a", all[0].ID)
}

func TestValidateAdd(t *testing.T) {
	t.Parallel()
	require.Error(t, api.ValidateAdd(api.AddRequest{Tray: "inbox", Title: "  "}))
	require.NoError(t, api.ValidateAdd(api.AddRequest{Tray: "inbox", Title: "Hello"}))
}

func ids(items []api.Item) []string {
	out := make([]string, len(items))
	for i := range items {
		out[i] = items[i].ID
	}
	return out
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/api/ -count=1 -race -run 'TestSnapshotFrom|TestFilterOpen|TestValidateAdd'`
Expected: FAIL with undefined `api.SnapshotFrom`

- [ ] **Step 3: Write minimal implementation**

`internal/api/dto.go`:

```go
package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/tjdsneto/tray-cli/internal/bar"
)

type ErrorBody struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type Session struct {
	SignedIn bool `json:"signedIn"`
}

type Snapshot struct {
	Badge        int       `json:"badge"`
	RemoteStatus string    `json:"remoteStatus"`
	Sections     []Section `json:"sections"`
}

type Section struct {
	Header string `json:"header"`
	Rows   []Row  `json:"rows"`
}

type Row struct {
	Scope     string `json:"scope"`
	TrayRef   string `json:"trayRef"`
	ItemID    string `json:"itemId"`
	Title     string `json:"title"`
	Clipboard string `json:"clipboard"`
}

type Tray struct {
	Scope     string `json:"scope"`
	Ref       string `json:"ref"`
	Name      string `json:"name"`
	OpenCount int    `json:"openCount"`
}

type Item struct {
	ID          string     `json:"id"`
	Scope       string     `json:"scope"`
	TrayRef     string     `json:"trayRef"`
	Title       string     `json:"title"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"createdAt"`
	SnoozeUntil *time.Time `json:"snoozeUntil,omitempty"`
	Body        *string    `json:"body,omitempty"`
	Actions     []string   `json:"actions"`
}

type AddRequest struct {
	Tray  string  `json:"tray"`
	Title string  `json:"title"`
	Body  *string `json:"body,omitempty"`
}

type DeclineRequest struct {
	Reason string `json:"reason"`
}

type SnoozeRequest struct {
	Until string `json:"until"`
}

func SnapshotFrom(snap bar.Snapshot) Snapshot {
	out := Snapshot{Badge: snap.Badge, RemoteStatus: snap.RemoteStatus, Sections: []Section{}}
	for _, sec := range snap.Sections {
		rows := make([]Row, 0, len(sec.Rows))
		for _, row := range sec.Rows {
			rows = append(rows, Row{
				Scope: string(row.Scope), TrayRef: row.TrayRef, ItemID: row.ItemID,
				Title: row.Title, Clipboard: bar.ClipboardLine(row),
			})
		}
		out.Sections = append(out.Sections, Section{Header: sec.Header, Rows: rows})
	}
	return out
}

func ValidateAdd(r AddRequest) error {
	if strings.TrimSpace(r.Tray) == "" || strings.TrimSpace(r.Title) == "" {
		return errInvalid("title and tray are required")
	}
	return nil
}

type invalidError string

func errInvalid(msg string) error { return invalidError(msg) }

func (e invalidError) Error() string { return string(e) }

func WriteError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorBody{Error: code, Message: message})
}
```

`internal/api/filter.go`:

```go
package api

import "sort"

func SelectItems(items []Item, filter string) []Item {
	out := make([]Item, 0, len(items))
	for _, it := range items {
		if filter == "all" || openWork(it) {
			out = append(out, it)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out
}

func openWork(it Item) bool {
	if it.Scope == "local" {
		return it.Status == "open"
	}
	switch it.Status {
	case "pending", "accepted", "snoozed":
		return true
	default:
		return false
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/api/ -count=1 -race -run 'TestSnapshotFrom|TestFilterOpen|TestValidateAdd'`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/api/dto.go internal/api/filter.go internal/api/dto_test.go
git commit -m "feat: shape the desktop API snapshot and item list"
```

---

### Task 4: HTTP handler, auth, session, and snapshot

**Files:**
- Create: `internal/api/source.go`
- Create: `internal/api/handler.go`
- Test: `internal/api/handler_test.go`

- [ ] **Step 1: Write the failing test**

```go
package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tjdsneto/tray-cli/internal/api"
	"github.com/tjdsneto/tray-cli/internal/bar"
)

type fakeSource struct {
	signedIn bool
	local    []bar.Item
	remote   []bar.Item
	remoteErr error
}

func (f fakeSource) SignedIn(api.Ctx) (bool, error) { return f.signedIn, nil }
func (f fakeSource) Attention(api.Ctx) ([]bar.Item, []bar.Item, error) {
	return f.local, f.remote, f.remoteErr
}
func (f fakeSource) Trays(api.Ctx) ([]api.Tray, error) { return nil, nil }
func (f fakeSource) Items(api.Ctx, string) ([]api.Item, error) { return nil, nil }
func (f fakeSource) Item(api.Ctx, string) (api.Item, error) {
	return api.Item{}, api.ErrNotFound
}
func (f fakeSource) Add(api.Ctx, api.AddRequest) (api.Item, error) {
	return api.Item{}, nil
}
func (f fakeSource) Act(api.Ctx, string, string, api.ActBody) error { return nil }

func TestSnapshot_requiresBearerAndKeepsLastRemote(t *testing.T) {
	t.Parallel()
	src := &fakeSource{signedIn: true, local: []bar.Item{{
		Scope: bar.ScopeLocal, TrayRef: "notes", TrayName: "notes", ItemID: "l1", Title: "L", CreatedAt: time.Unix(1, 0).UTC(),
	}}, remote: []bar.Item{{
		Scope: bar.ScopeRemote, TrayRef: "inbox", TrayName: "inbox", ItemID: "r1", Title: "R", CreatedAt: time.Unix(2, 0).UTC(),
	}}}
	h := api.NewHandler("secret", src, time.Hour)
	h.Rebuild(t.Context())

	denied := httptest.NewRecorder()
	h.ServeHTTP(denied, httptest.NewRequest(http.MethodGet, "/v1/snapshot", nil))
	require.Equal(t, http.StatusUnauthorized, denied.Code)

	src.remoteErr = api.ErrRemote
	h.Rebuild(t.Context())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/snapshot", nil)
	req.Header.Set("Authorization", "Bearer secret")
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	var snap api.Snapshot
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &snap))
	require.Equal(t, "remote unavailable", snap.RemoteStatus)
	require.Equal(t, 2, snap.Badge)
}
```

`t.Context()` needs Go 1.24+. This module is Go 1.25, so that call is valid. If a test helper is clearer, use `context.Background()` instead of `t.Context()` in both the test and later steps. Pick one and use it in every handler test.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/api/ -count=1 -race -run TestSnapshot_requiresBearerAndKeepsLastRemote`
Expected: FAIL with undefined `api.NewHandler`

- [ ] **Step 3: Write minimal implementation**

`internal/api/source.go`:

```go
package api

import (
	"context"
	"errors"

	"github.com/tjdsneto/tray-cli/internal/bar"
)

type Ctx = context.Context

var ErrNotFound = errors.New("not found")
var ErrRemote = errors.New("remote unavailable")
var ErrSignedOut = errors.New("signed out")

type ActBody struct {
	Reason string
	Until  string
}

type Source interface {
	SignedIn(Ctx) (bool, error)
	Attention(Ctx) (local []bar.Item, remote []bar.Item, remoteErr error)
	Trays(Ctx) ([]Tray, error)
	Items(Ctx, trayRef string) ([]Item, error)
	Item(Ctx, id string) (Item, error)
	Add(Ctx, AddRequest) (Item, error)
	Act(Ctx, id, action string, body ActBody) error
}
```

`internal/api/handler.go` for this task serves `GET /v1/session` and `GET /v1/snapshot` only. Later tasks add routes to the same `ServeHTTP` switch. Include this switch now:

```go
package api

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/tjdsneto/tray-cli/internal/bar"
)

type Handler struct {
	token    string
	src      Source
	interval time.Duration
	mu       sync.Mutex
	snap     bar.Snapshot
}

func NewHandler(token string, src Source, interval time.Duration) *Handler {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	return &Handler{token: token, src: src, interval: interval}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(r) {
		WriteError(w, http.StatusUnauthorized, "unavailable", "unauthorized")
		return
	}
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/v1/session":
		h.session(w, r)
	case r.Method == http.MethodGet && r.URL.Path == "/v1/snapshot":
		h.snapshot(w, r)
	default:
		WriteError(w, http.StatusNotFound, "not_found", "not found")
	}
}

func (h *Handler) authorized(r *http.Request) bool {
	const prefix = "Bearer "
	got := r.Header.Get("Authorization")
	if !strings.HasPrefix(got, prefix) {
		return false
	}
	token := got[len(prefix):]
	return subtle.ConstantTimeCompare([]byte(token), []byte(h.token)) == 1
}

func (h *Handler) session(w http.ResponseWriter, r *http.Request) {
	ok, err := h.src.SignedIn(r.Context())
	if err != nil {
		WriteError(w, http.StatusBadGateway, "unavailable", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, Session{SignedIn: ok})
}

func (h *Handler) snapshot(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	snap := SnapshotFrom(h.snap)
	h.mu.Unlock()
	writeJSON(w, http.StatusOK, snap)
}

func (h *Handler) Rebuild(ctx Ctx) {
	local, remote, remoteErr := h.src.Attention(ctx)
	h.mu.Lock()
	defer h.mu.Unlock()
	if remoteErr != nil {
		remote = bar.RemoteItemsFromSnapshot(h.snap)
		h.snap = bar.BuildSnapshot(append(append([]bar.Item{}, local...), remote...), "remote unavailable")
		return
	}
	h.snap = bar.BuildSnapshot(append(append([]bar.Item{}, local...), remote...), "")
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
```

Fix the test's `fakeSource` methods to pointer receivers if `src.remoteErr = ...` must be visible. The test holds `src := &fakeSource{}` and passes `src` to `NewHandler`. Methods must have pointer receivers. In Step 1, change every `func (f fakeSource)` to `func (f *fakeSource)` before running the test. `Attention` returns `f.remoteErr`.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/api/ -count=1 -race -run TestSnapshot_requiresBearerAndKeepsLastRemote`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/api/source.go internal/api/handler.go internal/api/handler_test.go
git commit -m "feat: serve the desktop snapshot on loopback"
```

---

### Task 5: Trays, item list, and item body

**Files:**
- Modify: `internal/api/handler.go`
- Modify: `internal/api/handler_test.go`

- [ ] **Step 1: Write the failing test**

Append to `handler_test.go`. Extend `fakeSource` with fields `trays []api.Tray` and `items []api.Item`, and implement `Trays` and `Items` by returning them. `Items` ignores the tray ref in this test and returns `f.items`. `Item` returns the matching id or `api.ErrNotFound`. List items must not include `Body`. `GET /v1/items/{id}` must include it.

```go
func TestItems_listOmitsBodyAndGetIncludesIt(t *testing.T) {
	t.Parallel()
	body := "full notes"
	src := &fakeSource{items: []api.Item{
		{ID: "c", Scope: "remote", TrayRef: "inbox", Title: "New", Status: "pending", CreatedAt: time.Unix(3, 0).UTC(), Body: &body, Actions: []string{"accept", "decline", "snooze"}},
		{ID: "b", Scope: "remote", TrayRef: "inbox", Title: "Old", Status: "completed", CreatedAt: time.Unix(1, 0).UTC()},
	}}
	h := api.NewHandler("secret", src, time.Hour)
	rec := getAuth(h, "/v1/items?tray=inbox&filter=open")
	require.Equal(t, http.StatusOK, rec.Code)
	var list []api.Item
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &list))
	require.Equal(t, []string{"c"}, ids(list))
	require.Nil(t, list[0].Body)

	one := getAuth(h, "/v1/items/c")
	require.Equal(t, http.StatusOK, one.Code)
	var item api.Item
	require.NoError(t, json.Unmarshal(one.Body.Bytes(), &item))
	require.Equal(t, "full notes", *item.Body)

	missing := getAuth(h, "/v1/items/nope")
	require.Equal(t, http.StatusNotFound, missing.Code)
}

func getAuth(h http.Handler, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("Authorization", "Bearer secret")
	h.ServeHTTP(rec, req)
	return rec
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/api/ -count=1 -race -run TestItems_listOmitsBodyAndGetIncludesIt`
Expected: FAIL with 404

- [ ] **Step 3: Add the routes**

In `ServeHTTP`, before the default branch:

```go
case r.Method == http.MethodGet && r.URL.Path == "/v1/trays":
    h.trays(w, r)
case r.Method == http.MethodGet && r.URL.Path == "/v1/items":
    h.items(w, r)
case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/v1/items/"):
    h.item(w, r)
```

```go
func (h *Handler) trays(w http.ResponseWriter, r *http.Request) {
	trays, err := h.src.Trays(r.Context())
	if err != nil {
		writeSourceErr(w, err)
		return
	}
	if trays == nil {
		trays = []Tray{}
	}
	writeJSON(w, http.StatusOK, trays)
}

func (h *Handler) items(w http.ResponseWriter, r *http.Request) {
	tray := r.URL.Query().Get("tray")
	filter := r.URL.Query().Get("filter")
	if filter == "" {
		filter = "open"
	}
	items, err := h.src.Items(r.Context(), tray)
	if err != nil {
		writeSourceErr(w, err)
		return
	}
	out := SelectItems(items, filter)
	for i := range out {
		out[i].Body = nil
		if out[i].Actions == nil {
			out[i].Actions = ActionsFor(out[i].Scope, out[i].Status)
		}
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) item(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/v1/items/")
	if id == "" || strings.Contains(id, "/") {
		WriteError(w, http.StatusNotFound, "not_found", "not found")
		return
	}
	item, err := h.src.Item(r.Context(), id)
	if err != nil {
		writeSourceErr(w, err)
		return
	}
	if item.Actions == nil {
		item.Actions = ActionsFor(item.Scope, item.Status)
	}
	writeJSON(w, http.StatusOK, item)
}

func writeSourceErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		WriteError(w, http.StatusNotFound, "not_found", "not found")
	case errors.Is(err, ErrSignedOut):
		WriteError(w, http.StatusUnauthorized, "signed_out", "signed out")
	case errors.Is(err, ErrRemote):
		WriteError(w, http.StatusBadGateway, "remote_unavailable", "remote unavailable")
	default:
		var inv invalidError
		if errors.As(err, &inv) {
			WriteError(w, http.StatusBadRequest, "invalid", inv.Error())
			return
		}
		WriteError(w, http.StatusBadGateway, "unavailable", err.Error())
	}
}
```

Import `errors` in `handler.go`. Export nothing new. `invalidError` is already in package `api`.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/api/ -count=1 -race -run TestItems_listOmitsBodyAndGetIncludesIt`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/api/handler.go internal/api/handler_test.go
git commit -m "feat: list desktop trays and load an item body on open"
```

---

### Task 6: Writes leave the snapshot unchanged on failure

**Files:**
- Modify: `internal/api/handler.go`
- Modify: `internal/api/handler_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestAct_failureDoesNotRebuildSnapshot(t *testing.T) {
	t.Parallel()
	src := &fakeSource{local: []bar.Item{{
		Scope: bar.ScopeLocal, TrayRef: "notes", TrayName: "notes", ItemID: "l1", Title: "L", CreatedAt: time.Unix(1, 0).UTC(),
	}}}
	src.actErr = errors.New("nope")
	h := api.NewHandler("secret", src, time.Hour)
	h.Rebuild(t.Context())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/items/l1/complete", nil)
	req.Header.Set("Authorization", "Bearer secret")
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadGateway, rec.Code)

	snap := getAuth(h, "/v1/snapshot")
	var body api.Snapshot
	require.NoError(t, json.Unmarshal(snap.Body.Bytes(), &body))
	require.Equal(t, 1, body.Badge)
	require.Equal(t, "l1", body.Sections[0].Rows[0].ItemID)
}
```

Add `actErr error` to `fakeSource`. `Act` returns `f.actErr`. Use `context.Background()` if `t.Context` is not what the rest of the file uses.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/api/ -count=1 -race -run TestAct_failureDoesNotRebuildSnapshot`
Expected: FAIL with 404

- [ ] **Step 3: Add mutation routes**

```go
case r.Method == http.MethodPost && r.URL.Path == "/v1/items":
    h.add(w, r)
case r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/v1/items/"):
    h.act(w, r)
```

`item()` must stay on GET only. A POST path `/v1/items/{id}/complete` is handled by `act`, not `item`. Match POST before GET in the switch.

```go
func (h *Handler) add(w http.ResponseWriter, r *http.Request) {
	var req AddRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid", "invalid json")
		return
	}
	if err := ValidateAdd(req); err != nil {
		writeSourceErr(w, err)
		return
	}
	item, err := h.src.Add(r.Context(), req)
	if err != nil {
		writeSourceErr(w, err)
		return
	}
	h.Rebuild(r.Context())
	item.Actions = ActionsFor(item.Scope, item.Status)
	writeJSON(w, http.StatusCreated, item)
}

func (h *Handler) act(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(r.URL.Path, "/v1/items/")
	id, action, ok := strings.Cut(rest, "/")
	if !ok || id == "" || action == "" {
		WriteError(w, http.StatusNotFound, "not_found", "not found")
		return
	}
	var body ActBody
	if r.ContentLength != 0 {
		var raw struct {
			Reason string `json:"reason"`
			Until  string `json:"until"`
		}
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			WriteError(w, http.StatusBadRequest, "invalid", "invalid json")
			return
		}
		body.Reason, body.Until = raw.Reason, raw.Until
	}
	if err := h.src.Act(r.Context(), id, action, body); err != nil {
		writeSourceErr(w, err)
		return
	}
	h.Rebuild(r.Context())
	w.WriteHeader(http.StatusNoContent)
}
```

A failed `Act` returns before `Rebuild`. That is the behavior under test.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/api/ -count=1 -race`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/api/handler.go internal/api/handler_test.go
git commit -m "feat: reject a failed desktop write without changing the snapshot"
```

---

### Task 7: `tray api` command

**Files:**
- Create: `internal/cli/commands/api.go`
- Modify: `internal/cli/commands/register.go`
- Test: `internal/cli/commands/api_test.go`

- [ ] **Step 1: Write the failing test**

```go
package commands

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHandshakeLine(t *testing.T) {
	t.Parallel()
	line := handshakeLine(4321, "tok")
	var got struct {
		Port  int    `json:"port"`
		Token string `json:"token"`
	}
	require.NoError(t, json.Unmarshal([]byte(line), &got))
	require.Equal(t, 4321, got.Port)
	require.Equal(t, "tok", got.Token)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/cli/commands/ -count=1 -race -run TestHandshakeLine`
Expected: FAIL with `undefined: handshakeLine`

- [ ] **Step 3: Write the command**

```go
package commands

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/tjdsneto/tray-cli/internal/api"
)

func cmdAPI() *cobra.Command {
	return &cobra.Command{
		Use:   "api",
		Short: "Serve the local desktop API on 127.0.0.1",
		RunE:  runAPI,
	}
}

func runAPI(cmd *cobra.Command, _ []string) error {
	token, err := randomToken()
	if err != nil {
		return err
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	port := ln.Addr().(*net.TCPAddr).Port
	fmt.Fprintln(cmd.OutOrStdout(), handshakeLine(port, token))

	h := api.NewHandler(token, newAPISource(cmdDeps.ConfigDir, cmdDeps.RequireAuth), 30*time.Second)
	h.Rebuild(cmd.Context())
	srv := &http.Server{Handler: h}
	go h.Loop(cmd.Context())

	ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		_, _ = io.Copy(io.Discard, cmd.InOrStdin())
		_ = srv.Shutdown(context.Background())
	}()
	go func() {
		<-ctx.Done()
		_ = srv.Shutdown(context.Background())
	}()
	err = srv.Serve(ln)
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func handshakeLine(port int, token string) string {
	b, _ := json.Marshal(struct {
		Port  int    `json:"port"`
		Token string `json:"token"`
	}{Port: port, Token: token})
	return string(b)
}

func randomToken() (string, error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}
```

Add `Loop` to `internal/api/handler.go`:

```go
func (h *Handler) Loop(ctx Ctx) {
	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			h.Rebuild(ctx)
		}
	}
}
```

In `register.go`, next to `cmdBar()`:

```go
root.AddCommand(cmdAPI())
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/cli/commands/ -count=1 -race -run TestHandshakeLine`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/cli/commands/api.go internal/cli/commands/api_test.go internal/cli/commands/register.go internal/api/handler.go
git commit -m "feat: start tray api as a loopback child process"
```

---

### Task 8: Real source over local files and the existing services

**Files:**
- Create: `internal/cli/commands/api_source.go`
- Test: `internal/cli/commands/api_source_test.go`

- [ ] **Step 1: Write the failing test**

Use a temp config dir and `localtray.NewStore`. Add a global tray item, list it through the source, complete it, and assert the open list is empty. This is the local-files half. Remote methods return `api.ErrSignedOut` when `RequireAuth` fails. Do not call the network.

```go
func TestAPISource_localAddListAndComplete(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := newAPISource(func() string { return dir }, func() (domain.Services, domain.Session, error) {
		return domain.Services{}, domain.Session{}, errors.New("signed out")
	})
	item, err := src.Add(context.Background(), api.AddRequest{Tray: "notes", Title: "Draft"})
	require.NoError(t, err)
	require.Equal(t, "local", item.Scope)
	require.Equal(t, "open", item.Status)

	items, err := src.Items(context.Background(), "notes")
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Nil(t, items[0].Body)

	err = src.Act(context.Background(), item.ID, "complete", api.ActBody{})
	require.NoError(t, err)
	items, err = src.Items(context.Background(), "notes")
	require.NoError(t, err)
	require.Empty(t, api.SelectItems(items, "open"))

	err = src.Act(context.Background(), item.ID, "accept", api.ActBody{})
	require.Error(t, err)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/cli/commands/ -count=1 -race -run TestAPISource_localAddListAndComplete`
Expected: FAIL with `undefined: newAPISource`

- [ ] **Step 3: Implement the source**

Add to `internal/api/dto.go`:

```go
func Invalid(msg string) error { return invalidError(msg) }
```

`internal/cli/commands/api_source.go`:

```go
func newAPISource(configDir func() string, auth func() (domain.Services, domain.Session, error)) *apiSource {
	return &apiSource{configDir: configDir, auth: auth}
}

func (s *apiSource) Add(_ api.Ctx, req api.AddRequest) (api.Item, error) {
	if err := api.ValidateAdd(req); err != nil {
		return api.Item{}, err
	}
	store := localtray.NewStore(s.configDir())
	target := localtray.AddTarget{Kind: localtray.KindGlobal, Name: strings.TrimSpace(req.Tray)}
	if rec, ok := findLocal(store, req.Tray); ok {
		target = localtray.AddTarget{Kind: rec.Kind, Name: rec.Name, Path: rec.Path, RepoRoot: rec.RepoRoot, Branch: rec.Branch}
	}
	item, rec, err := store.AddItem(target, strings.TrimSpace(req.Title), true, req.Body)
	if err != nil {
		return api.Item{}, err
	}
	return fromLocal(item, rec), nil
}

func (s *apiSource) Items(_ api.Ctx, trayRef string) ([]api.Item, error) {
	store := localtray.NewStore(s.configDir())
	rec, ok := findLocal(store, trayRef)
	if !ok {
		return s.remoteItems(trayRef)
	}
	rows, err := store.ListItemsForTrays([]string{rec.ID})
	if err != nil {
		return nil, err
	}
	out := make([]api.Item, 0, len(rows))
	for _, row := range rows {
		out = append(out, fromLocal(row.Item, row.Tray))
	}
	return out, nil
}

func (s *apiSource) Act(_ api.Ctx, id, action string, body api.ActBody) error {
	store := localtray.NewStore(s.configDir())
	if _, err := findLocalItem(store, id); err == nil {
		if action != "complete" {
			return api.Invalid("local items only support complete")
		}
		_, err = store.CompleteItem(id)
		return err
	}
	svcs, sess, err := s.auth()
	if err != nil {
		return api.ErrSignedOut
	}
	return remoteAct(svcs, sess, id, action, body)
}
```

`findLocal` loads the index and returns the record whose `bar.LocalTrayRef` equals the ref. `fromLocal` sets `Scope: "local"`, `TrayRef: bar.LocalTrayRef(rec)`, `Status: item.Status`, `Actions: api.ActionsFor("local", item.Status)`, and copies `Body`. `findLocalItem` scans `ListItemsForTrays(localtray.AllTrayIDs(idx))`.

`remoteAct` uses `svcs.Items.Update` with status `accepted`, `declined`, `completed`, `archived`, or `snoozed`. Decline sets `DeclineReason` when the reason is non-empty. Snooze parses `body.Until` with `time.RFC3339` then `time.RFC3339Nano`, rejects a time that is not after `time.Now()`, and sets `SnoozeUntil`. Any other action returns `api.Invalid("unknown action")`.

`Attention` returns `bar.ItemsFromLocal(store.AllOpenItems())`. Auth failure returns those local items, nil remote, and nil error. Auth success lists owned trays and pending items (`ListItemsQuery{Status: "pending", OrderCreated: "desc", TrayIDIn: ids}`) and maps them with `bar.ItemsFromRemote` and `trayref.TrayNameMap`. A remote list error returns `api.ErrRemote`. `SignedIn` is auth error == nil. `Trays` uses `Store.TraySummaries` for local `OpenCount`. Remote trays are omitted when signed out. Remote `OpenCount` counts items whose status is pending, accepted, or snoozed.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/cli/commands/ -count=1 -race -run TestAPISource_localAddListAndComplete`
Expected: PASS

Then: `go test ./internal/api/ ./internal/cli/commands/ -count=1 -race`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/cli/commands/api_source.go internal/cli/commands/api_source_test.go internal/api/dto.go
git commit -m "feat: back the desktop API with local trays and the existing services"
```

Task 8 Step 3 is the one place the code is described against existing functions rather than pasted in full, because `TargetFromID` and `AddItem` must be read in the files named above before the call is written. The test in Step 1 is the contract. Do not add remote network fixtures.

---

### Task 9: User docs

**Files:**
- Create: `docs/user/desktop.md`
- Modify: `docs/user/README.md`
- Modify: `README.md`

- [ ] **Step 1: Write the page**

Create `docs/user/desktop.md` with this body:

```markdown
# Desktop app

The Tray desktop app is a window for trays you own, plus the macOS menu bar.

## Launch

The app runs the `tray` binary on your `PATH` and talks to it on this machine only. Install the CLI first. Sign in with `tray login` before remote trays appear. Local trays work signed out.

`tray bar` still works. Running it at the same time as the app shows two menu bar icons.

## Window

Attention is the same set as the menu bar badge. A tray shows open work, newest first. Completed and archived items sit behind All. Opening a row loads its body. Buttons that do not apply to that item are not shown.

## Menu bar

The icon's menu still copies `tray-ref item-id` when you click a row. Open shows the window.
```

Add a link to that page from `docs/user/README.md` next to the menu bar link, and one sentence plus the same link in the root `README.md` after the `tray bar` mention if one exists. If the root README does not mention the menu bar, add the desktop sentence under "First use".

- [ ] **Step 2: Commit**

```bash
git add docs/user/desktop.md docs/user/README.md README.md
git commit -m "docs: describe the tray desktop app"
```

---

### Task 10: Desktop scaffold

**Files:**
- Create: `desktop/package.json`
- Create: `desktop/vite.config.ts`
- Create: `desktop/index.html`
- Create: `desktop/tsconfig.json`
- Create: `desktop/src/main.tsx`
- Create: `desktop/src/styles.css`
- Create: `desktop/src-tauri/Cargo.toml`
- Create: `desktop/src-tauri/tauri.conf.json`
- Create: `desktop/src-tauri/build.rs`
- Create: `desktop/src-tauri/src/main.rs`
- Create: `desktop/src-tauri/capabilities/default.json`

- [ ] **Step 1: Write the frontend package**

`desktop/package.json` scripts: `"dev": "vite"`, `"build": "vite build"`, `"test": "vitest run"`. Dependencies: `react`, `react-dom`, `@tauri-apps/api`. Dev dependencies: `vite`, `@vitejs/plugin-react`, `typescript`, `vitest`, `jsdom`, `@testing-library/react`. Tailwind is a dependency of the visual spec. Add `tailwindcss`, `postcss`, `autoprefixer`, and a `desktop/tailwind.config.js` whose `content` is `./index.html` and `./src/**/*.{ts,tsx}`. `styles.css` starts with the three Tailwind directives and:

```css
.reading { transition: opacity 150ms ease; }
```

`vite.config.ts` sets `server.port` to `1420` and `server.strictPort` to true, matching GoFlow. `clearScreen: false`.

`index.html` mounts `<div id="root">` and loads `/src/main.tsx`. `main.tsx` renders `<App />` from `./App`. Create `desktop/src/App.tsx` that returns `<h1>Tray</h1>` so the shell builds before Task 12 replaces it.

- [ ] **Step 2: Write the Tauri shell**

`desktop/src-tauri/Cargo.toml` uses `tauri = { version = "2", features = ["macos-private-api", "tray-icon"] }`, `tauri-plugin-single-instance = "2"`, `serde`, `serde_json`. Package name `tray-desktop`. `build.rs` is `fn main() { tauri_build::build() }`.

`tauri.conf.json` product name `Tray`, identifier `com.tjdsneto.tray`, `beforeDevCommand` `npm run dev`, `devUrl` `http://localhost:1420`, `frontendDist` `../dist`. One window labeled `main`, title `Tray`, width 1100, height 720, `visible` true.

`capabilities/default.json` allows the main window core permissions the default Tauri 2 template grants (`core:default`).

`src/main.rs` for this task only builds a window. Task 13 adds the child process and the menu. Plugin:

```rust
.plugin(tauri_plugin_single_instance::init(|app, _args, _cwd| {
    if let Some(window) = app.get_webview_window("main") {
        let _ = window.show();
        let _ = window.set_focus();
    }
}))
```

A second launch focuses that window.

- [ ] **Step 3: Install and build the frontend**

Run from `desktop/`: `npm install` then `npm test -- --passWithNoTests` and `npm run build`.
Expected: build writes `desktop/dist`. Do not start `tauri dev` in this task.

- [ ] **Step 4: Commit**

```bash
git add desktop
git commit -m "feat: scaffold the tray desktop shell"
```

---

### Task 11: Action buttons

**Files:**
- Create: `desktop/src/actions.ts`
- Create: `desktop/src/ActionBar.tsx`
- Test: `desktop/src/ActionBar.test.tsx`
- Create: `desktop/src/snooze.ts`
- Test: `desktop/src/snooze.test.ts`

- [ ] **Step 1: Write the failing tests**

`ActionBar.test.tsx` renders `<ActionBar actions={...} onAction={() => {}} />`.

- `["accept", "decline", "snooze"]` shows buttons Accept, Decline, Snooze and does not show Complete.
- `["complete", "archive"]` shows Complete and Archive.
- `["complete"]` shows only Complete.

`snooze.test.ts` uses `new Date("2026-06-15T18:30:00-03:00")`. `snoozeUntil("hour", now)` is one hour later. `snoozeUntil("tomorrow", now)` is `2026-06-16T09:00:00-03:00`. `snoozeUntil("week", now)` is `2026-06-22T09:00:00-03:00`.

- [ ] **Step 2: Run test to verify it fails**

Run from `desktop/`: `npx vitest run src/ActionBar.test.tsx src/snooze.test.ts`
Expected: FAIL because the modules do not exist

- [ ] **Step 3: Implement**

`actions.ts` maps `accept`, `decline`, `complete`, `snooze`, `archive` to the labels Accept, Decline, Complete, Snooze, Archive. `ActionBar` renders one `<button type="button">` per action in that order and calls `onAction(action)`.

`snooze.ts`:

```ts
export function snoozeUntil(preset: "hour" | "tomorrow" | "week", now: Date): Date {
  if (preset === "hour") return new Date(now.getTime() + 60 * 60 * 1000);
  const next = new Date(now);
  next.setDate(now.getDate() + (preset === "tomorrow" ? 1 : 7));
  next.setHours(9, 0, 0, 0);
  return next;
}
```

The reading pane sends `snoozeUntil(preset, new Date()).toISOString()` as `until`. It does not send a preset name.

- [ ] **Step 4: Run test to verify it passes**

Run: `npx vitest run src/ActionBar.test.tsx src/snooze.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add desktop/src/actions.ts desktop/src/ActionBar.tsx desktop/src/ActionBar.test.tsx desktop/src/snooze.ts desktop/src/snooze.test.ts
git commit -m "feat: show only the desktop actions an item allows"
```

---

### Task 12: Three panes

**Files:**
- Create: `desktop/src/api.ts`
- Create: `desktop/src/App.tsx`
- Modify: `desktop/src/styles.css`

- [ ] **Step 1: Write the client**

`api.ts` exports `connect(base: string, token: string)` returning functions `session`, `snapshot`, `trays`, `items(tray, filter)`, `item(id)`, `add`, `act(id, action, body?)`. Every call sends `Authorization: Bearer`. On a 401, invoke the Tauri command `restart_api` once, replace `base` and `token` with its result, and retry the request. A second 401 throws the error JSON `message`. `items` requests `filter=open` unless the caller passes `all`.

The Rust shell from Task 13 will set `window.__TRAY_API__ = { base, token }` before the webview loads. Until that exists, `App` reads that global and shows "tray is not installed" when it is missing.

- [ ] **Step 2: Write the window**

`App.tsx` state:

```ts
type Selection = { kind: "attention" } | { kind: "tray"; ref: string };
```

Left pane: a button Attention with `snapshot.badge`, then trays grouped by `scope` `local` then `remote`. Signed out (`session.signedIn === false`) renders one muted line "Sign in with tray login" instead of remote trays. `remoteStatus === "remote unavailable"` renders that line and still shows the last remote trays.

Middle pane: when selection is attention, render `snapshot.sections` (header, then rows). No add field. When selection is a tray, render `items` newest-first as returned by the API, a text field "Add to {name}", and a control that toggles `filter` between `open` and `all`. Submit posts `add` and refetches.

Right pane: selecting a row fetches `item(id)` and renders title, status, body, and `<ActionBar actions={item.actions}>`. The pane has class `reading`. Decline shows an optional reason input and posts `{ reason }`. Snooze shows three buttons, Hour, Tomorrow, Next week, and posts `{ until }` from `snoozeUntil`. Complete and Archive post with an empty body. After a successful act, refetch the snapshot and the open list. A failed act leaves the list in state and shows `message` from the error JSON.

Poll `snapshot` and, when a tray is selected, `items`, every 30 seconds.

- [ ] **Step 3: Check the frontend tests still pass**

Run from `desktop/`: `npx vitest run`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add desktop/src
git commit -m "feat: add the three-pane tray window"
```

---

### Task 13: Menu bar in the Tauri process

**Files:**
- Modify: `desktop/src-tauri/src/main.rs`
- Modify: `desktop/src-tauri/Cargo.toml` if `arboard` is not already there. Add `arboard = "3"` for the clipboard. The row copies text. Tauri does not need the Go clipboard package.

- [ ] **Step 1: Spawn `tray api`**

Before the window is shown, `Command::new("tray").arg("api")` with piped stdin and stdout. Read one line. Parse `{"port","token"}`. On spawn failure, still open the window and evaluate JS that leaves `__TRAY_API__` unset. On success, inject:

```rust
window.eval(&format!(
    "window.__TRAY_API__ = {{ base: \"http://127.0.0.1:{port}\", token: \"{token}\" }};"
))
```

Hold the `Child` and its stdin in Tauri state so stdin stays open. On `RunEvent::Exit`, drop stdin and kill the child.

Register a command `restart_api`. It kills the current child, calls the same spawn function, stores the new child, and returns `{ "base": "http://127.0.0.1:<port>", "token": "<token>" }`. That is what the webview invokes after one 401.

- [ ] **Step 2: Build the status menu from the snapshot**

Every 30 seconds, `GET /v1/snapshot` with the bearer token. Rebuild the tray menu. Each row is `title` and its id is the `clipboard` string. On click, `arboard::Clipboard::new()?.set_text(clipboard)`. A fixed item **Open** shows and focuses the `main` window. Set the tray tooltip to the badge count. Badge `0` uses the tooltip `Tray` and no number.

This menu is macOS because that is where `tray-icon` is exercised for v1. The window still opens on other OS. Do not remove `tray bar`.

- [ ] **Step 3: Commit**

```bash
git add desktop/src-tauri
git commit -m "feat: own the menu bar from the desktop app"
```

---

## Spec coverage

| Spec | Task |
|------|------|
| Action table, including local open and no snooze on accepted | 1, 11 |
| Snooze presets at 09:00 local | 2, 11 |
| Shared snapshot, last remote rows, 30s refresh | 4, 7 |
| Body loaded on open, list without body | 5, 12 |
| Failed write leaves the snapshot | 6 |
| Handshake, loopback, stdin EOF, SIGTERM, bearer token | 7 |
| Local files and existing domain services | 8 |
| Signed out remote line, remote unavailable | 8, 12 |
| Docs, two icons, `tray bar` stays | 9 |
| Three panes, Attention first, add only on a tray, newest first, All filter | 3, 12 |
| Fade the reading pane | 10, 12 |
| Optional decline reason, no completion message | 12 |
| Menu bar copy and Open, second launch focuses | 10, 13 |
| Missing `tray` on PATH | 13 |
| 401 restarts the child once | 12, 13 |
| No Tauri end-to-end test | not a task |

## Self-review

- Placeholder scan: no TBD, TODO, or "similar to task N". Task 8's remote list and tray counts name the existing `ListOwned`, `List`, and `TraySummaries` calls. The local add, list, and complete path is code.
- Type names used throughout: `ActionsFor`, `SnoozeUntil`, `SnapshotFrom`, `SelectItems`, `ValidateAdd`, `NewHandler`, `Source.Act`, `handshakeLine`, `newAPISource`.
- The UI does not reimplement the action table. It renders `item.actions` from Go.
