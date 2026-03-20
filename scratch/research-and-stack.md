# Research & Stack: Tray CLI

This document summarizes similar products, how Tray CLI (attention-queue repo) differs from them, the **chosen** stack, and follow-up work tracked in bd.

### Stack decisions (2026-03-20)

| Piece | Choice |
|-------|--------|
| **Product / tool name** | **Tray CLI** (binary: `tray`) |
| **CLI implementation** | **Go** (e.g. cobra for commands) |
| **Backend** | **Supabase** — Postgres, Auth, RLS, PostgREST (Data API), optional Realtime / Edge Functions |

---

## 1. Similar Projects

### CLI-first / queue-style task tools

| Tool | Description | Shared / multi-user? | Link |
|------|-------------|----------------------|------|
| **tqu** | Minimal queue-based task tracking; push/pop workflow, multiple queues (e.g. "daily", "bills"), SQLite, Python, MIT | No (local only) | [primaprashant/tqu](https://github.com/primaprashant/tqu) |
| **nq** | Unix CLI job queue; timestamp-ordered jobs, filesystem sync via `flock`, `$NQDIR` for shared dir = shared queue; no daemon | Yes (via shared dir) | [nq](https://git.vuxu.org/nq/about/) |
| **task-q** | Linux CLI for multi-user queue; shared task management across users | Yes | [rodra-go/task-q](https://github.com/rodra-go/task-q) |
| **queue (toddlehman)** | Parallel job scheduler, per-user DB at `~/.queue`, priority (a–z), `add` from stdin | No | [toddlehman/queue](https://github.com/toddlehman/queue) |
| **MyQueue** | Task queue with `submit`, dependencies, resources; CLI doc at [MyQueue CLI](https://myqueue.readthedocs.io/cli.html) | Context-dependent | MyQueue docs |

### Agent / delegation-oriented

| Tool | Description | Shared / “inbox from others”? | Link |
|------|-------------|-------------------------------|------|
| **Delega** | Task API for AI agents: REST + MCP + CLI; agent identity, delegation chains, assign_to, webhooks, semantic dedup; cloud or self-host, MIT | Yes (assign to agent/user) | [delega.dev](https://delega.dev/), [GitHub](https://github.com/delega-dev/delega) |
| **Beans** | CLI-first issue tracker, markdown files, TUI, GraphQL for agents, team collaboration | Yes (team) | [hmans/beans](https://github.com/hmans/beans) |
| **Agentic Work Queue** | VS Code extension + CLI (`wq-cli.js`), worklist from markdown, status/deps; for Claude Code | No (local workflow) | [fasutron/vscode-agentic-work-queue](https://github.com/fasutron/vscode-agentic-work-queue) |

### Shared inbox / assign-to-me (productivity)

| Product | Description | CLI-first? |
|---------|-------------|------------|
| **Linear** | Assign issues to people; Inbox for “assigned to you” and updates | No (web/API) |
| **Asana** | Shared projects, assign tasks; Gmail/Polymail → tasks | No |
| **Front** | Shared inboxes, assign conversations to teammates | No |
| **Todoist** | Shared projects, collaborate with others; assign in team plans | No (mobile/web; API exists) |

### “Non-urgent” / triage-adjacent

- **RADAR** (research): Agent-assisted email → task list; task-centric vs inbox-centric. Not a shipping product.
- **Email triage** (Martin, ClearContext, etc.): Triage/label inbox, some delegation; not a dedicated “attention queue” from others.
- **ClearContext Delegate**: Assign emails to colleagues; creates reminder tasks. Tied to email, not a standalone queue.

**Summary:** There are local CLI queues (tqu, nq, task-q), agent task APIs with assign/delegate (Delega), and team task/inbox products (Linear, Asana, Front). None are clearly positioned as a **personal, multi-life-area, shared “put it on my radar” queue** with **invite-based members** and **CLI-first + agent-friendly** as the main interface.

---

## 2. How Tray CLI Would Differ

| Dimension | Typical tools | Tray CLI |
|-----------|----------------|-----------------|
| **Primary use** | Team tasks, agent jobs, or local queues | Personal “attention queue” others can add to (work, life, other buckets) |
| **Who adds items** | Self, or same-org assignees | Invited/accepted people (“put on my radar”) without full task ownership |
| **Urgency** | Often mixed with urgent/assignments | Explicitly non-urgent; dedicated channel so items aren’t lost in email/chat |
| **Scope** | Usually work or single domain | Configurable buckets across life areas |
| **Interface** | Often web-first or IDE-first | CLI-first and agent-friendly (terminal + Claude/agents) |
| **Complexity** | Full task/project/assignee model | Lighter: queue + review + prioritization rules, not full project management |

So the product is meaningfully different: **lightweight, personal, shared “radar” queue, multi-bucket, CLI/agent-first**, rather than “another task manager” or “another shared inbox.”

---

## 3. Stack (CLI-first) — Tray CLI

Requirement: **Tray CLI** as the primary interface, usable from terminal and from agents (e.g. Claude), with a backend that supports multi-tenant, auth, and shared trays (buckets).

### Backend (API + auth + multi-tenant)

- **Option A – Supabase (recommended for MVP)**  
  - **Why:** Postgres, built-in auth (email, OAuth, magic links), RLS for tenant isolation, realtime optional, REST and client SDKs. Fast to ship; CLI can talk to API via REST or SDK.  
  - **Multi-tenant:** One “tenant” = one user’s trays; RLS by `user_id` (and optional `bucket`). “Contributors” are extra roles or a join table (user_id, owner_id, role).  
  - **Auth:** Supabase Auth; CLI uses session + refresh or API key (e.g. service role or user JWT) for server-side; for CLI, device/session or API key per user is typical.

- **Option B – Custom backend (Node/Go/Rust + Postgres)**  
  - **Why:** Full control over API shape and auth (e.g. API keys for CLI/agents). More work: auth, RLS or app-level tenancy, and hosting.  
  - **When:** If you need very specific semantics (e.g. custom delegation model) or want to avoid Supabase lock-in.

**Recommendation:** Start with **Supabase** (Postgres + Auth + RLS). Add a small **API layer** (e.g. Edge Functions or a minimal Node/Go service) only if you need endpoints that don’t map cleanly to Supabase (e.g. custom “invite” or “review” semantics).

### Data store

- **Primary:** **PostgreSQL** (via Supabase).  
  - Tables: users (or use Supabase Auth), queues (or one queue per user), items (title, bucket, source_user_id, priority, status, created_at), bucket_members (owner_id, user_id, permissions), and optionally prioritization_rules.  
- **Caching/sessions:** Optional **Redis** (or Supabase Realtime) later for presence or rate limiting; not required for MVP.

### CLI (Tray CLI)

- **Language (decided):** **Go** — single static binary, easy install (`brew install`, curl script), strong HTTP/JSON stdlib and ecosystem.  
- **Responsibilities:**  
  - Auth: login (e.g. browser or device flow), store token under `~/.config/tray/` (or keychain); see [cli-design.md](cli-design.md#local-storage-local-db).  
  - Commands: e.g. `tray add`, `tray list`, `tray review`, `tray invite`, **`tray listen`** (watch for new/updated items; see *Listen / watch* below) — full set in [cli-design.md](cli-design.md).  
  - All operations via **REST API** (or Supabase client) so the same API can serve a future web UI or MCP server.  
- **Agent-friendly:** Stable, documented commands; JSON output flag (e.g. `tray list --json`) for parsing; optional **MCP server** that wraps the same API for Claude/Cursor.

**Implementation:** **Go** + **cobra** (or similar) for subcommands and flags. Implement **`--json`** on all read commands and document the schema.

### Listen / watch (CLI)

**Requirement:** Users and agents should be able to **listen** for activity on their queue(s)—new items, updates—and react (e.g. output to terminal, run a command, pipe elsewhere).

- **Modes:**
  - **Poll:** `tray listen` (or `tray watch`) polls the API on an interval (e.g. every N seconds); simple, works everywhere, no long-lived connection.
  - **Realtime (optional):** Use Supabase Realtime (or WebSocket) so the CLI gets pushed events; lower latency, fewer API calls, but more complexity (connection handling, reconnects).
- **Output:**
  - **Default:** Stream events to stdout (one line or one JSON object per event), so users can `tray listen` and see items as they arrive, or `tray listen --json` for machine-readable.
  - **Wire to something else:** e.g. `tray listen --exec 'notify-send "New item: {}"'`, or `tray listen --json | jq -r '.payload.title' | xargs -I {} my-script "{}"`. Avoid building too much into the CLI; keep it composable (Unix-style).
- **Scope:** Listen to all buckets or a specific tray (`tray listen [tray]`); optionally filter by event type (new item, status change, etc.) via flags.

**Recommendation for MVP:** Implement **poll-based** `tray listen` with configurable interval (e.g. `--interval 30s`), stdout output (plain and `--json`), and optional `--exec <cmd>` so the CLI stays simple and scriptable. Add Realtime later if needed.

### Queue vs stack (semantics)

The product is an **attention queue**: one ordered list of items others add to. The open question is **consumption order**: FIFO (queue) vs LIFO (stack).

- **Queue (FIFO):** Process in order added—"oldest first." Matches the name and the "inbox / radar" metaphor: work through the backlog.
- **Stack (LIFO):** Process newest first—"what just landed." Some users prefer to triage recent items before older ones.

**Recommendation:** Offer **one list** (the queue) but support **both orderings as a view**:

- **Default:** Queue order (FIFO) for `tray list`, `tray review`, and any "next item" semantics. This keeps the product clearly a *queue*.
- **Optional:** A flag or config for "newest first" (stack-style view), e.g. `tray list --newest-first` or `list.order: newest_first` in config. Same data; only the order in which items are presented changes.

So: **queue** is the primary abstraction and naming; **stack** is an optional way to *view* the same queue (newest first). No need for two separate data structures or product concepts—just ordering preference on list/review.

### Optional: MCP server

- Expose “add to queue”, “list queue”, “review” as MCP tools so agents can use the same backend without parsing CLI output. Can be a thin wrapper around the same HTTP API.

### CLI distribution

**Goal:** Easiest possible install so users can run `tray` with minimal friction. A **single static Go binary** avoids runtime installs and works the same everywhere.

| Method | Effort | Best for | Notes |
|--------|--------|----------|--------|
| **One-liner install script** | Low | Everyone | `curl -sSL https://.../install.sh \| sh` — detect OS/arch, download binary, put in `$PATH`. Same pattern as rustup, deno, etc. Easiest to document ("copy-paste this"). |
| **Homebrew** | Low | macOS / Linux | `brew install <org>/tray/tray` (tap) or submit to core. Most devs on Mac already have it; Linux has homebrew too. |
| **Scoop / Winget** | Low | Windows | Scoop: add bucket, `scoop install tray`. Winget: publish to Microsoft or community repo. Covers Windows without asking users to install Go. |
| **Go install** | None | Contributors / power users | `go install github.com/<org>/tray@latest` — not the primary path for end users. |
| **GitHub Releases** | None | Manual / CI | Attach `tray-darwin-arm64`, `tray-linux-amd64`, `tray-windows-amd64.exe` etc. to releases. Install script and package managers pull from here. |

**Recommendation for MVP:**

1. **Ship a single Go binary** per OS/arch and publish to **GitHub Releases**.
2. **Provide an install script** that detects OS/arch, downloads the right binary, and places it in `~/bin` or a dir in `$PATH`. One URL in the README: "Run this to install."
3. **Add a Homebrew tap** (macOS/Linux): e.g. a Formula in `attention-queue/homebrew-tray` pointing at release assets; users run `brew install <tap>/tray`.
4. **Optional:** Scoop bucket for Windows so `scoop install tray` works.

Avoid for MVP: building and maintaining platform-specific packages (`.deb`, `.rpm`, `.msi`, Chocolatey, apt repos) unless there is a clear need. The script + Homebrew + GitHub Releases cover the vast majority of CLI users with minimal upkeep.

**Public vs private repo:** Anonymous install scripts and Homebrew need **public** download URLs. **Private** repos do not expose release assets to the world; **GitHub Pages** on a **free** account requires a **public** repo (or a separate public site repo). See **[repo-visibility-and-distribution.md](repo-visibility-and-distribution.md)** for details and pros/cons.

### Summary

| Layer | Technology | Rationale |
|-------|------------|-----------|
| API + Auth + Multi-tenant | **Supabase** (Postgres + Auth + RLS) | Fast MVP, auth and RLS out of the box; API key or JWT for CLI |
| Data store | **PostgreSQL** (Supabase) | Structured data, RLS, and future realtime if needed |
| Tray CLI | **Go + cobra**, optional **MCP** | Single binary, `--json`, agent-friendly; MCP for native agent use |
| **Distribution** | **GitHub Releases + install script + Homebrew tap** | One-liner install; brew for Mac/Linux; optional Scoop for Windows |
| Hosting | Supabase Cloud (or self-host Supabase) | No infra to run for MVP |

**Detailed specs:** [cli-design.md](cli-design.md) (CLI commands, remotes, local storage); [backend-spec.md](backend-spec.md) (API, auth, DB structure).

---

## 4. CLI API Design & Backend Implications

This section defines a concrete CLI surface and maps each part to backend requirements, then calls out challenges that affect backend choice.

### Proposed CLI surface

Canonical command names and flags live in **[cli-design.md](cli-design.md)**. Summary:

| Command | Purpose | Example |
|--------|---------|--------|
| `tray login` | Authenticate (browser/device flow), store token | `tray login` |
| `tray add "title" [tray-or-alias]` | Add item (owner: tray name; member: remote alias) | `tray add "Review deck" work` |
| `tray list [tray]` | List items, FIFO or newest-first | `tray list --json --newest-first` |
| `tray review [tray]` | Interactive or batch review | `tray review` |
| `tray prioritize` | Set priority or rules (post-MVP nuance) | `tray prioritize <id> high` |
| `tray invite <tray>` | Shareable invite link/token for a tray | `tray invite work` |
| `tray listen [tray]` | Watch for new/updated items (poll or realtime) | `tray listen --interval 30s --exec 'notify "{}"'` |
| `tray create` / `tray ls` | Create or list trays | see cli-design |

**Output:** All read commands support `--json`. List/review support `--newest-first` (stack view).

For triage (accept / decline / snooze / complete / archive), member view (“items I added”), bucket members/revoke/rotate-invite, and shareable-invite (join) flow, see **[Invite flow & scenarios](invite-flow-and-scenarios.md)**.

### Backend requirements per command

| CLI command | Backend need | Supabase | Firebase / custom |
|-------------|--------------|----------|-------------------|
| **login** | Issue and refresh tokens; store session. | Auth + JWT; CLI stores in `~/.config/...`. | Same (Auth). |
| **add** | Insert item with `bucket_id`, `source_user_id`; enforce "caller is owner or member for this bucket". | RLS: allow INSERT if `auth.uid()` = bucket owner OR in bucket_members. | Security rules: same idea. |
| **list** | Query items by bucket (or all my buckets), order by `created_at` ASC/DESC. Enforce: only items in buckets I own or am a member of. | RLS on `items` joined to `buckets` / `bucket_members`. | Firestore rules + query. |
| **review** | Read items + update `status` (or similar). Same visibility as list. | SELECT + UPDATE with RLS. | Read + write with rules. |
| **prioritize** | Update item priority or read/write prioritization_rules. | Extra column or table + RLS. | Field or subcollection + rules. |
| **invite** | Create pending invite, send email/link, **or** shareable invite token (self-join); accept/join flow creates bucket_members row. **Not plain CRUD.** See [Invite flow & scenarios](invite-flow-and-scenarios.md). | Edge Function or external service (e.g. send email, accept endpoint); token + join endpoint for shareable link. | Cloud Function + email; same for token. **Custom logic on both.** |
| **listen** | Poll: "items created/updated since T" (cursor or timestamp). Realtime: subscribe to inserts/updates. | Poll: `created_at > $1` or `updated_at > $1`. Realtime: Supabase Realtime on `items`. | Poll: query with filter. Realtime: Firestore listeners. |
| **bucket ls/create** | CRUD on buckets; only owner can create; list only my buckets. | RLS by `owner_id`. | Rules by owner. |

### Challenges the API poses for the backend

1. **Invite flow (custom logic)**  
   `tray invite` / email-invite flows imply: create invite record, send email (or magic link), and an "accept" path that creates a member and optionally triggers auth. That is **not** a single CRUD call.  
   - **Implication:** Whatever backend we pick must support **custom endpoints or server-side functions** (Supabase Edge Functions, Firebase Cloud Functions, or a small API service). Supabase and Firebase are even here; a "DB-only" backend is not enough.

2. **Contributor-based visibility and writes**  
   Owner sees all items in their buckets; members can add items and may need to see (e.g.) items in buckets they contribute to. So we need **row-level rules** that depend on a join: "user is owner of bucket" or "user is a member of that bucket".  
   - **Implication:** Supabase RLS with a policy that references `buckets` and `bucket_members` is a good fit. Firebase security rules can express the same with subcollections/denormalized "member_ids". No deal-breaker for either; RLS is a bit more explicit for "owner vs member" in SQL.

3. **Listen: poll vs realtime**  
   Poll needs an efficient "items since timestamp/cursor" query (index on `created_at` or `updated_at`). Realtime needs push of insert/update events.  
   - **Implication:** Supabase: indexed query + optional Realtime. Firebase: Firestore query + listeners. Both support both. Realtime is nicer for `tray listen` long-running; poll is simpler for MVP and works with any backend that supports filtered queries.

4. **CLI talks REST vs SDK**  
   If the CLI uses the **Supabase client** (or Firebase SDK), it is tied to that backend. If the CLI talks **REST** to a thin API layer, we can swap the backend behind the API.  
   - **Implication:** Prefer **REST (or OpenAPI)** for the CLI's server communication so the backend is an implementation detail. Then Supabase can sit behind a small API, or we can use Supabase PostgREST directly if we accept that the CLI is "Supabase-aware" (PostgREST is still REST). Same for Firebase (REST via Firebase REST API or a custom API).

### Conclusion: does the CLI API block or favor a backend?

- **No blocking issues.** The proposed CLI can be implemented with Supabase, Firebase, or a custom backend. All of them need custom code for **invite**; the rest is CRUD + auth + visibility rules.
- **Supabase** remains a good fit: Postgres + RLS models "owner vs member" clearly; Realtime and polling both work; Edge Functions cover invite. The main constraint is committing to Postgres and Supabase's auth/model.
- **Firebase** works if we prefer NoSQL and Firestore rules; we'd need Cloud Functions for invite and possibly a small API if we want a single REST surface for the CLI.
- **Recommendation:** Keep **Supabase** as the recommended backend. The CLI API does not introduce requirements that Supabase can't meet; the only non-CRUD piece (invite) requires custom logic on any backend. See **[Invite flow, item lifecycle & scenarios](invite-flow-and-scenarios.md)** for invite models (owner-invite vs shareable link), token rotation, member view (“items I added”), and item statuses (accept/decline/snooze/complete/archive, due date). Document the invite and join flows in the Supabase spike.

---

## 5. Follow-up Work (bd issues)

The following bd issues were created for spikes and documentation. (Created with `BEADS_DIR=~/.beads`; if this repo gets a local beads DB via `bd init`, consider re-creating or linking these there.)

| ID | Title |
|----|--------|
| `tjdsneto-9tj` | Spike: evaluate Supabase for Tray CLI backend |
| `tjdsneto-668` | Spike: evaluate Delega as baseline or integration |
| `tjdsneto-guc` | Document: competitor matrix |
| `tjdsneto-klp` | ~~Spike: CLI tech~~ **Done:** Go (Tray CLI) |

- **Supabase spike** – Confirm RLS schema for user + bucket_members + buckets and auth flow for CLI (API key vs JWT).
- **Delega spike** – Compare Delega’s assign/delegate model and MCP/CLI with our “members add to my queue” model; decide build vs integrate.
- **Competitor matrix** – One-pager comparing tqu, nq, task-q, Delega, Linear/Asana/Front on: shared queue, CLI, multi-tenant, agent-friendly, “radar” positioning.
- **CLI tech** – **Go + cobra** chosen for Tray CLI; distribution strategy is in §3 (GitHub Releases + install script + Homebrew tap).

Run `bd ready` (with `BEADS_DIR=~/.beads` if using global beads) to see unblocked work, and `bd show <id>` for full descriptions.

---

*Last updated: 2026-03-20 (Tray CLI, Go + Supabase decisions; `tray` binary; cli-design alignment).*
