# Agent session trays Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship remote per-agent-session inbox trays (`agent-session:<id>`), create-on-add, source `agent_session_id` stamping (GoFlow-aligned flags/env), list/prune, hooks + docs/skill awareness.

**Architecture:** Reserved remote tray **name** `agent-session:<id>` (no `trays.kind` column). Pure helpers in `internal/cli/agentsession` for name/parse/prune eligibility. CLI `--agent-session-id` ensures/creates the tray then adds; item column `agent_session_id` stores the **source** session. Prune extends `tray prune --remote` for empty session trays idle ≥ 7 days. Dial/hotline ping stays skill-only.

**Tech Stack:** Go 1.25, cobra, PostgREST/Supabase, existing `trayref` / `ItemService` / `TrayService`, testify.

**Spec:** [`docs/superpowers/specs/2026-09-24-agent-session-trays-design.md`](../specs/2026-09-24-agent-session-trays-design.md)

**Locked decisions:** name prefix (not kind column); 7-day prune idle; create-on-add under writer’s account only; no listen “mine” filter in v1.

---

## File map

| Path | Role |
|------|------|
| `internal/cli/agentsession/name.go` | Pure: `TrayName`, `ParseTrayName`, `NormalizeID`, prefix const |
| `internal/cli/agentsession/prune.go` | Pure: `IsOpenStatus`, `Prunable` (empty-of-open + idle) |
| `internal/cli/agentsession/*_test.go` | Unit tests |
| `supabase/migrations/20260925120000_items_agent_session_id.sql` | `items.agent_session_id text` |
| `internal/domain/item.go` | `AgentSessionID *string`; widen `Add` |
| `internal/adapters/postgrest/item.go` | `itemRow`, `addItemRequest`, `newAddItemRequest` |
| `internal/adapters/postgrest/item_service.go` | `Add` + `itemSelectColumns` |
| `internal/adapters/postgrest/item_service_test.go` | Add body / select tests |
| `internal/output/items.go` (+ test) | JSON (+ optional table) field |
| `internal/cli/listenhook/env.go` (+ test) | `TRAY_ITEM_AGENT_SESSION_ID` |
| `internal/cli/commands/items_run.go` | Flags, ensure tray, stamp source |
| `internal/cli/commands/agentsession_ensure.go` | Thin ensure/create helper for remote |
| `internal/cli/commands/local_prune.go` | Extend prune for `--remote` session trays |
| `docs/user/hooks.md`, `docs/user/trays.md` or new short section | User docs |
| `skills/tray-cli/SKILL.md`, `CLAUDE.md`, `AGENTS.md` | Agent awareness |

---

### Task 1: Pure `agentsession` name helpers

**Files:**
- Create: `internal/cli/agentsession/name.go`
- Create: `internal/cli/agentsession/name_test.go`

- [ ] **Step 1: Write the failing test**

```go
package agentsession_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tjdsneto/tray-cli/internal/cli/agentsession"
)

func TestTrayName_and_Parse(t *testing.T) {
	t.Parallel()
	name, err := agentsession.TrayName("  abc-123  ")
	require.NoError(t, err)
	require.Equal(t, "agent-session:abc-123", name)

	id, ok := agentsession.ParseTrayName(name)
	require.True(t, ok)
	require.Equal(t, "abc-123", id)

	_, err = agentsession.TrayName("")
	require.Error(t, err)
	_, err = agentsession.TrayName("   ")
	require.Error(t, err)

	_, ok = agentsession.ParseTrayName("inbox")
	require.False(t, ok)
	_, ok = agentsession.ParseTrayName("agent-session:")
	require.False(t, ok)
}

func TestIsAgentSessionTrayName(t *testing.T) {
	t.Parallel()
	require.True(t, agentsession.IsAgentSessionTrayName("agent-session:x"))
	require.False(t, agentsession.IsAgentSessionTrayName("Agent-Session:x")) // stored names are exact prefix
	require.False(t, agentsession.IsAgentSessionTrayName("work"))
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/cli/agentsession/ -count=1`

Expected: FAIL (package does not exist)

- [ ] **Step 3: Write minimal implementation**

```go
package agentsession

import (
	"fmt"
	"strings"
)

const NamePrefix = "agent-session:"

func NormalizeID(id string) (string, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return "", fmt.Errorf("agent session id is empty")
	}
	return id, nil
}

func TrayName(id string) (string, error) {
	id, err := NormalizeID(id)
	if err != nil {
		return "", err
	}
	return NamePrefix + id, nil
}

func ParseTrayName(name string) (id string, ok bool) {
	name = strings.TrimSpace(name)
	if !strings.HasPrefix(name, NamePrefix) {
		return "", false
	}
	id = strings.TrimSpace(strings.TrimPrefix(name, NamePrefix))
	if id == "" {
		return "", false
	}
	return id, true
}

func IsAgentSessionTrayName(name string) bool {
	_, ok := ParseTrayName(name)
	return ok
}
```

- [ ] **Step 4: Run tests — expect PASS**

Run: `go test ./internal/cli/agentsession/ -count=1`

- [ ] **Step 5: Commit**

```bash
git add internal/cli/agentsession/
git commit -m "$(cat <<'EOF'
Add agentsession tray name helpers.

EOF
)"
```

---

### Task 2: Pure prune eligibility

**Files:**
- Create: `internal/cli/agentsession/prune.go`
- Create: `internal/cli/agentsession/prune_test.go`

- [ ] **Step 1: Write the failing test**

```go
package agentsession_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tjdsneto/tray-cli/internal/cli/agentsession"
	"github.com/tjdsneto/tray-cli/internal/domain"
)

func TestPrunable(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)
	idle := 7 * 24 * time.Hour
	old := now.Add(-8 * 24 * time.Hour)
	recent := now.Add(-2 * 24 * time.Hour)

	tray := domain.Tray{Name: "agent-session:abc", UpdatedAt: old}
	require.True(t, agentsession.Prunable(tray, nil, now, idle), "empty + idle")

	tray.UpdatedAt = recent
	require.False(t, agentsession.Prunable(tray, nil, now, idle), "empty but not idle")

	tray.UpdatedAt = old
	open := []domain.Item{{Status: "accepted", UpdatedAt: old}}
	require.False(t, agentsession.Prunable(tray, open, now, idle), "has open item")

	done := []domain.Item{{Status: "completed", UpdatedAt: old, CompletedAt: &old}}
	require.True(t, agentsession.Prunable(tray, done, now, idle), "only completed + idle")

	require.False(t, agentsession.Prunable(domain.Tray{Name: "inbox", UpdatedAt: old}, nil, now, idle))
}

func TestIsOpenStatus(t *testing.T) {
	t.Parallel()
	require.True(t, agentsession.IsOpenStatus("pending"))
	require.True(t, agentsession.IsOpenStatus("accepted"))
	require.True(t, agentsession.IsOpenStatus("snoozed"))
	require.False(t, agentsession.IsOpenStatus("completed"))
	require.False(t, agentsession.IsOpenStatus("declined"))
	require.False(t, agentsession.IsOpenStatus("archived"))
}
```

- [ ] **Step 2: Run — expect FAIL** (`Prunable` undefined)

- [ ] **Step 3: Implement**

```go
package agentsession

import (
	"strings"
	"time"

	"github.com/tjdsneto/tray-cli/internal/domain"
)

const DefaultPruneIdle = 7 * 24 * time.Hour

func IsOpenStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "pending", "accepted", "snoozed":
		return true
	default:
		return false
	}
}

// Prunable reports whether an agent-session tray may be deleted: correct name,
// no open items, and last activity (tray.UpdatedAt or latest item UpdatedAt) older than idle.
func Prunable(tray domain.Tray, items []domain.Item, now time.Time, idle time.Duration) bool {
	if !IsAgentSessionTrayName(tray.Name) {
		return false
	}
	last := tray.UpdatedAt
	for _, it := range items {
		if IsOpenStatus(it.Status) {
			return false
		}
		if it.UpdatedAt.After(last) {
			last = it.UpdatedAt
		}
	}
	if idle <= 0 {
		idle = DefaultPruneIdle
	}
	return !last.IsZero() && now.Sub(last) >= idle
}
```

- [ ] **Step 4: Run — expect PASS**

Run: `go test ./internal/cli/agentsession/ -count=1`

- [ ] **Step 5: Commit**

```bash
git add internal/cli/agentsession/
git commit -m "$(cat <<'EOF'
Add agentsession prune eligibility helper.

EOF
)"
```

---

### Task 3: Migration + domain `AgentSessionID`

**Files:**
- Create: `supabase/migrations/20260925120000_items_agent_session_id.sql`
- Modify: `internal/domain/item.go`
- Modify: `internal/domain/item.go` — `ItemService.Add` signature

- [ ] **Step 1: Add migration**

```sql
-- Source AI-agent session id that created the item (GF-37-aligned; optional).

alter table public.items
  add column if not exists agent_session_id text;

comment on column public.items.agent_session_id is
  'Optional originating AI-agent session id (source). Target inbox is the agent-session:<id> tray name.';
```

- [ ] **Step 2: Extend domain**

On `Item` struct add:

```go
AgentSessionID *string
```

Change interface to:

```go
Add(ctx context.Context, sess Session, trayID, title string, dueDate *string, agentSessionID *string) (*Item, error)
```

(Only call site today: `items_run.go` — update in Task 5; temporarily pass `nil` if compiling mid-way.)

- [ ] **Step 3: Commit**

```bash
git add supabase/migrations/20260925120000_items_agent_session_id.sql internal/domain/item.go
git commit -m "$(cat <<'EOF'
Add items.agent_session_id column and domain field.

EOF
)"
```

---

### Task 4: PostgREST map + `Add` body

**Files:**
- Modify: `internal/adapters/postgrest/item.go`
- Modify: `internal/adapters/postgrest/item_service.go`
- Modify: `internal/adapters/postgrest/item_service_test.go` (extend Add test)

- [ ] **Step 1: Write/extend failing test**

In `item_service_test.go`, extend `TestItemService_Add_ownerGetsAccepted` (or add sibling) to assert POST body includes `"agent_session_id":"sess-a"` when non-nil is passed. Inspect how existing tests capture request body and mirror that pattern.

- [ ] **Step 2: Run — expect FAIL**

- [ ] **Step 3: Implement**

1. `itemRow`: add `AgentSessionID *string \`json:"agent_session_id"\`` and map in `ToDomain`.
2. `itemSelectColumns`: append `agent_session_id`.
3. `addItemRequest`: add `AgentSessionID *string \`json:"agent_session_id,omitempty"\``.
4. `newAddItemRequest(..., agentSessionID *string)` — if trimmed non-empty, set on request.
5. `itemService.Add`: accept `agentSessionID *string`, pass through.

- [ ] **Step 4: Run**

Run: `go test ./internal/adapters/postgrest/ -count=1 -run AgentSession`

(or full package if filter name differs)

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/adapters/postgrest/ internal/domain/item.go
git commit -m "$(cat <<'EOF'
Wire agent_session_id through PostgREST item create/list.

EOF
)"
```

---

### Task 5: Output + hooks env

**Files:**
- Modify: `internal/output/items.go`
- Modify: `internal/output/items_test.go`
- Modify: `internal/cli/listenhook/env.go`
- Modify: `internal/cli/listenhook/listenhook_test.go`

- [ ] **Step 1: Failing tests**

JSON row must include `agent_session_id` when set (omitempty or null — match existing pointer style in `WriteItems`).

Hooks: `HookEnv` includes `TRAY_ITEM_AGENT_SESSION_ID=<id>` when `it.AgentSessionID` set; omit when nil.

Const name: `EnvItemAgentSessionID = "TRAY_ITEM_AGENT_SESSION_ID"` (do **not** reuse `TRAY_AGENT_SESSION_ID`, which is the live agent’s identity for “mine”).

- [ ] **Step 2: Implement minimal wiring in `WriteItems` and `HookEnv`**

- [ ] **Step 3: Run**

```bash
go test ./internal/output/ ./internal/cli/listenhook/ -count=1
```

Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/output/ internal/cli/listenhook/
git commit -m "$(cat <<'EOF'
Surface agent_session_id in item JSON and hook env.

EOF
)"
```

---

### Task 6: Ensure remote session tray + `tray add` flags

**Files:**
- Create: `internal/cli/commands/agentsession_ensure.go`
- Create: `internal/cli/commands/agentsession_ensure_test.go` (pure parts if any; otherwise test via `flagOrEnv` helper)
- Modify: `internal/cli/commands/items_run.go`

- [ ] **Step 1: Pure helper for flag/env (TDD)**

Add in `agentsession` or commands:

```go
// FlagOrEnv returns trimmed flag if non-empty, else trimmed getenv(key), else "".
func FlagOrEnv(flagVal, envKey string) string
```

Test empty flag falls back to env; flag wins; both empty → "".

(GoFlow pattern: `flag_or_env`.)

- [ ] **Step 2: Ensure helper (I/O — keep thin)**

```go
// ensureAgentSessionTray returns tray id for agent-session:<id>, creating if missing under sess.
func ensureAgentSessionTray(ctx context.Context, svcs *app.Services, sess domain.Session, agentSessionID string) (string, error)
```

Logic:
1. `name, err := agentsession.TrayName(agentSessionID)`
2. `owned, err := svcs.Trays.ListMine(...)` or owned-only list used elsewhere for create-on-add — use the same list as `tray create` / owned trays (see `runCreate` / `ListOwned` if present; prefer owned-only so names don’t collide with joined trays incorrectly).
3. `matches := trayref.FindTraysByNameFold(owned, name)` → if one, return ID; if many, error.
4. If none: `svcs.Trays.Create(ctx, sess, name, nil)` → return new ID.
5. On unique-violation race: list again and pick.

Inspect existing tray list method names in `domain.TrayService` and use the owned-trays API.

- [ ] **Step 3: Wire `cmdAdd`**

Flags:
- `String("agent-session-id", "", "remote inbox for this AI agent session id (implies --remote)")`
- `String("from-agent-session-id", "", "source agent session id stamped on the item (default $TRAY_AGENT_SESSION_ID)")`

Behavior in `runAdd` / `runRemoteItemAdd`:
- If `--agent-session-id` set (or env used only for list “mine”, **not** as implicit add target): require auth; force remote path; ignore positional tray name for target; `ensureAgentSessionTray`; source := `FlagOrEnv(fromFlag, "TRAY_AGENT_SESSION_ID")`; pass `*source` or nil to `Items.Add`.
- Mutual exclusion: `--agent-session-id` with `--here`/`--branch`/`--project` → error.
- Empty target id → validation error.

When `--agent-session-id` is set, do not require a positional tray argument.

- [ ] **Step 4: Manual smoke (optional in plan execution)**

```bash
export TRAY_AGENT_SESSION_ID=test-sess-a
tray add "hello from A" --agent-session-id test-sess-b --remote
tray list --agent-session-id test-sess-b --remote --format json
```

- [ ] **Step 5: Commit**

```bash
git add internal/cli/commands/ internal/cli/agentsession/
git commit -m "$(cat <<'EOF'
Add --agent-session-id create-on-add for remote inboxes.

EOF
)"
```

---

### Task 7: `tray list --agent-session-id`

**Files:**
- Modify: `internal/cli/commands/items_run.go` (`cmdList` / `runRemoteItemList`)

- [ ] **Step 1: Flags on `cmdList`**

Same `--agent-session-id` flag:
- With value → resolve/ensure-not-required for list: look up owned (and visible) tray by `agent-session:<id>`; if missing, print empty list (or friendly “no tray yet”) — **do not create on list**.
- Flag present with empty value → use `$TRAY_AGENT_SESSION_ID`; if still empty, error: pass id or set env.
- Implies `--remote`.

- [ ] **Step 2: Unit-test any new pure resolve-for-list helper** (name → optional tray id from slice).

- [ ] **Step 3: `go test ./internal/cli/... -count=1`**

- [ ] **Step 4: Commit**

```bash
git add internal/cli/commands/items_run.go
git commit -m "$(cat <<'EOF'
List remote agent-session inbox via --agent-session-id.

EOF
)"
```

---

### Task 8: `tray prune --remote` for idle session trays

**Files:**
- Modify: `internal/cli/commands/local_prune.go` (or split `prune.go` if file grows)

- [ ] **Step 1: Extend command**

```text
tray prune              # local empty dir/branch (unchanged)
tray prune --remote     # remote agent-session:* trays that are Prunable
tray prune --remote --dry-run
```

Optional: `--idle 168h` defaulting to `agentsession.DefaultPruneIdle`.

Update `Short`/`Long` to mention both.

- [ ] **Step 2: Remote prune flow**

1. Auth required.
2. List owned trays; filter `IsAgentSessionTrayName`.
3. For each: `Items.List` for that tray id (all statuses or enough to judge open).
4. If `Prunable(tray, items, now, idle)` → `Trays.Delete` (find existing delete API — same as `tray remove` remote path).
5. Print pruned names like local prune.

- [ ] **Step 3: Test pure eligibility already covered; add command-level test only if there is an existing prune test harness — otherwise manual dry-run.**

- [ ] **Step 4: Commit**

```bash
git add internal/cli/commands/local_prune.go
git commit -m "$(cat <<'EOF'
Prune idle empty remote agent-session trays.

EOF
)"
```

---

### Task 9: Docs + skill + instruction mirrors

**Files:**
- Modify: `docs/user/hooks.md` — document `TRAY_ITEM_AGENT_SESSION_ID`
- Modify: `docs/user/trays.md` or `docs/user/local-trays.md` — short “Agent session trays” section pointing at remote `--agent-session-id`
- Modify: `skills/tray-cli/SKILL.md` — awareness: export `TRAY_AGENT_SESSION_ID`, check inbox, optional dial ping after add
- Modify: `CLAUDE.md` / `AGENTS.md` — one short bullet mirroring skill (parity rule)
- Modify: `docs/user/README.md` link if new doc file

- [ ] **Step 1: Write docs/skill text** (no code behavior change)

Skill bullets (canonical ideas):
- At session start, `export TRAY_AGENT_SESSION_ID=<your session id>` (same id as agent-registry).
- Check inbox: `tray list --agent-session-id --remote` (uses env).
- Hand off: `tray add "…" --agent-session-id <other> --remote` (stamps source from env).
- Optional: dial/hotline the other session that an item is waiting — not required.
- Prune: `tray prune --remote` removes idle empty session trays; add recreates.

- [ ] **Step 2: Commit**

```bash
git add docs/ skills/tray-cli/SKILL.md CLAUDE.md AGENTS.md
git commit -m "$(cat <<'EOF'
Document agent session trays for users and agents.

EOF
)"
```

---

## Spec coverage checklist

| Spec requirement | Task |
|------------------|------|
| Remote tray per session `agent-session:<id>` | 1, 6 |
| Create-on-add / recreate after prune | 6 |
| Source `agent_session_id` + GoFlow naming | 3, 4, 5, 6 |
| List mine / by id | 7 |
| Prune empty idle (7d) | 2, 8 |
| Hooks expose source id | 5 |
| Skill + CLAUDE/AGENTS awareness | 9 |
| No new ACL / no listen mine filter | (explicit non-work) |

---

## Execution handoff

Plan complete and saved to `docs/superpowers/plans/2026-09-24-agent-session-trays.md`.

**Two execution options:**

1. **Subagent-Driven (recommended)** — fresh subagent per task, review between tasks  
2. **Inline Execution** — run tasks in this session with executing-plans checkpoints  

Which approach?
