# Launch Requirements: MVP

What we need for a **working MVP** for **Tray CLI** (Go + Supabase) that people can install and use: one person owns a tray (bucket), shares it via a link, someone else adds items, the owner triages, and the contributor sees status. This doc is the single checklist for "ready to launch"; detailed design lives in [research-and-stack.md](research-and-stack.md), [invite-flow-and-scenarios.md](invite-flow-and-scenarios.md), [cli-design.md](cli-design.md), and [backend-spec.md](backend-spec.md).

---

## 1. MVP in one sentence

**A user can install the CLI, sign up, create a bucket, share a link; a colleague can join via that link and add items; the owner can list, triage (accept/decline/snooze/complete), and the colleague can see the status of what they added.**

---

## 2. User journey (must work end-to-end)

| Step | Who | Action | Outcome |
|------|-----|--------|---------|
| 1 | New user | Installs CLI (install script or Homebrew) | `tray` works in terminal |
| 2 | New user | Runs `tray login` | Authenticated; token stored |
| 3 | Owner | Creates a bucket, gets shareable invite | e.g. `tray create work` → `tray invite work` prints link |
| 4 | Colleague | Opens invite link (or uses token in CLI), signs up/logs in, joins | Becomes member; can add items to that bucket (e.g. via remote alias) |
| 5 | Colleague | Adds an item to owner’s bucket | `tray add "Review deck when you have time" <alias>` → item in queue |
| 6 | Owner | Lists and triages items | `tray list work`, then accept / decline / snooze / complete (CLI or `tray review`) |
| 7 | Colleague | Checks status of what they added | `tray contributed` shows items and status (pending, accepted, completed, declined) |

If this flow works for one owner and one contributor, the MVP is **usable**.

---

## 3. Requirements by area

### 3.1 Distribution (install)

| Requirement | Notes |
|-------------|--------|
| Single binary per OS/arch | Go build; no runtime required on user machine. |
| GitHub Releases | Binaries for at least: darwin/arm64, darwin/amd64, linux/amd64, windows/amd64. |
| Install script | One-liner (e.g. `curl -sSL ... \| sh`) that detects OS/arch, downloads binary, adds to `$PATH`. |
| Homebrew tap (optional for day-one) | Formula in a tap repo pointing at release URL; `brew install …` works. |

**Repo visibility:** Public Releases = anyone can install without a GitHub login. Private repo = need another public host for binaries (or a small public “releases” repo). GitHub Pages on a **free** plan needs a **public** repo for the Pages source (or host the auth success page elsewhere). See **[repo-visibility-and-distribution.md](repo-visibility-and-distribution.md)**.

**Done when:** A new user can run one command (or download from Releases) and get `tray` in their path.

---

### 3.2 Auth

| Requirement | Notes |
|-------------|--------|
| Sign up / sign in | Email+password or OAuth (e.g. GitHub/Google); Supabase Auth covers this. |
| CLI login | `tray login` opens browser (or device flow); token stored in `~/.config/tray/` (or equivalent). |
| Token usage | CLI sends token with API requests; backend validates and identifies user. |

**Done when:** User can `tray login` once and subsequent commands are authenticated.

#### Do we need a web page? Signup through CLI?

- **Auth always involves a browser** for the actual login (OAuth redirect or magic-link click). So you need *some* URL where the browser lands after success.
- You do **not** need a full “signup web app” or custom forms. Options:
  - **OAuth (GitHub / Google):** User clicks “Sign in with GitHub” in the browser; signup is on the provider. You only need a **callback/success page** that the redirect hits — e.g. “You’re signed in. Return to the CLI.” or a page that passes the token back to the CLI (e.g. localhost callback or “copy this token”).
  - **Supabase Hosted Auth:** Supabase can host the login/signup UI; you set a **redirect URL** to your success page (or a minimal page you host).
- **Signup “through the CLI”** means: user runs `tray login` → CLI opens browser (or prints URL + code for device flow) → user completes auth in browser → CLI gets the token (polling or callback). So the flow is CLI-initiated; the browser is only for the provider’s consent screen and your success/callback page. No need for a separate “go to our website to sign up” step.

**MVP:** One minimal **success/callback page** (can be Supabase’s or a single static page). No full web app required.

#### Do we need email sending?

- **Magic-link auth:** Yes — user enters email, you send a link. Supabase can send via their default or your SMTP.
- **OAuth-only (GitHub / Google):** **No email sending.** User signs up by “Sign in with GitHub”; no verification or transactional email needed from your side.
- **Invite (MVP):** We use shareable link only; no email invites. So no invite emails.

**MVP:** Use **OAuth-only** (e.g. GitHub + Google) and **no magic link**. Then you need **zero email sending** and no SMTP setup. Add magic link + email later if you want.

#### Do we need our own domain?

- **No.** You can run entirely on Supabase’s domain: e.g. `https://<project>.supabase.co` for API and auth. Configure Supabase auth redirect URL to point to a success page (could be Supabase’s or a free static host, e.g. GitHub Pages under a repo or Supabase hosted).
- **Custom domain** (e.g. `app.attentionqueue.com`) is for branding and nicer links; add when you want a polished product. Not required to launch.

**MVP:** Start **without** a custom domain. Use Supabase project URL + OAuth; add a domain when you care about branding or a cleaner invite/redirect URL.

---

### 3.3 Backend (API + data)

| Requirement | Notes |
|-------------|--------|
| Supabase project | Postgres + Auth + RLS; tables for users (Auth), buckets, items, bucket_members. |
| Buckets | Owner can create bucket; bucket has name, owner_id, invite_token (for shareable link). |
| Items | title, bucket_id, source_user_id, status, created_at; optional due_date, snooze_until, decline_reason, completion_message. |
| Bucket members | Join table: bucket_id, user_id; members can add items and read “their” items (source_user_id). |
| RLS | Owner: full CRUD on own buckets and their items. Member: add items to joined buckets; read items where source_user_id = self. |
| Invite/join | Join endpoint: accept invite_token → create bucket_member; optional Edge Function if not pure PostgREST. |

**Done when:** API supports create bucket, join by token, add item, list items (by bucket / by “my contributed”), update item status.

---

### 3.4 Invite flow (MVP: shareable link only)

| Requirement | Notes |
|-------------|--------|
| Shareable invite per bucket | Owner runs e.g. `tray invite <tray>`; gets URL (or token) to share. |
| Join by link/token | New or existing user uses link (or `tray join <url/token>`); backend creates bucket_member. |
| Remotes (aliases) | Member stores alias → bucket (e.g. in `~/.config/tray/remotes`); `tray add "..." <alias>` resolves to that bucket. |

**Out of scope for MVP:** Email-based invite (Model A). Can add later.

**Done when:** Owner can share a link; colleague can join and add items using an alias.

---

### 3.5 CLI: core commands

| Command | Purpose | MVP required |
|---------|---------|----------------|
| `tray login` | Authenticate, store token | Yes |
| `tray create <name>` | Create tray (bucket) | Yes |
| `tray ls` | List my trays | Yes |
| `tray invite <tray>` | Show/generate shareable invite link | Yes |
| `tray join <url-or-token>` | Join bucket via shareable invite; prompt for alias | Yes |
| `tray remote add <alias> <url-or-token>` | Add remote for a joined bucket | Yes (or implicit in join) |
| `tray remote ls` | List remotes | Yes |
| `tray add "title" [bucket-or-alias]` | Add item to own bucket or remote | Yes |
| `tray list [bucket]` | List items (default: active only), support `--json` | Yes |
| `tray accept <id>`, `tray decline <id>`, `tray snooze <id> [--until DATE]`, `tray complete <id>`, `tray archive <id>` | Triage actions | Yes |
| `tray contributed` | List items I added to others’ buckets, with status; `--json` | Yes |
| `tray review [bucket]` | Interactive triage (optional) | Nice-to-have |

**Done when:** All “Yes” commands work and are documented (e.g. `tray --help` and README).

---

### 3.6 Item lifecycle (MVP)

| Status | Meaning |
|--------|--------|
| `pending` | New; default for new items. |
| `accepted` | Owner acknowledged. |
| `declined` | Owner declined (optional reason). |
| `snoozed` | Deferred until date (optional snooze_until). |
| `completed` | Done (optional completion message). |
| `archived` | Dismissed without completing. |

Default list view: active = pending, accepted, snoozed (if snooze_until passed or null). Exclude declined/completed/archived from default view (or separate filter).

**Done when:** Owner can set these statuses; member sees them in `tray contributed`.

---

### 3.7 Agent-friendly

| Requirement | Notes |
|-------------|--------|
| `--json` on read commands | `tray list --json`, `tray contributed --json` for parsing. |
| Stable output schema | Document field names so agents can rely on them. |

**Done when:** All list/read commands support `--json` and schema is documented.

---

### 3.8 Out of scope for MVP

- **Model A (email invite)** — invite by email with approval flow.
- **Prioritization rules** — custom rules for ordering; manual ordering or default FIFO is enough.
- **`tray listen`** — poll or realtime watch; can add after launch.
- **MCP server** — optional later; CLI + `--json` is enough for agents at first.
- **Web UI** — CLI-only for MVP.
- **Multiple roles/permissions** — owner vs member is enough; no fine-grained RBAC.

---

## 4. Launch checklist

Use this to decide “we can launch.”

- [ ] **Install** — Install script works on macOS and Linux; binary available for Windows.
- [ ] **Auth** — `tray login` works; token persisted and used for API calls.
- [ ] **Buckets** — Create bucket, list buckets, generate shareable invite.
- [ ] **Join** — New user can join via invite link; remote/alias works so they can add items.
- [ ] **Add** — Owner and member can add items to the right bucket (owner by name, member by alias).
- [ ] **List** — Owner sees items in bucket(s); default view is active items only.
- [ ] **Triage** — Accept, decline, snooze, complete, archive; status persists and is visible.
- [ ] **Contributed** — Member can run `tray contributed` and see status of items they added.
- [ ] **JSON** — `tray list --json` and `tray contributed --json` work; schema documented.
- [ ] **Docs** — README with install, login, create bucket, share link, join, add, list, triage, contributed.
- [ ] **Error handling** — Clear errors for not logged in, invalid bucket/alias, missing id; no silent failures.

When all of the above are done, the MVP is **launchable**: someone can use it as their “attention queue” and share it with one other person. Iterate from there (email invite, listen, MCP, prioritization rules, etc.).

---

*References: [research-and-stack.md](research-and-stack.md), [invite-flow-and-scenarios.md](invite-flow-and-scenarios.md), [cli-design.md](cli-design.md), [backend-spec.md](backend-spec.md), [PRODUCT.md](PRODUCT.md).*

*Last updated: 2026-03-20 (Tray CLI, Go + Supabase; `tray` commands aligned with cli-design).*
