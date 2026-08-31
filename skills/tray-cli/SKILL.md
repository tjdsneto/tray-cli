---
name: tray-cli
description: >-
  Helps people use the **tray** CLI for **inboxes** (**trays**), **items**, **handoffs**,
  **triage**, and **attention** on queued work — **local context trays** (directory, branch,
  global) and **remote** trays for collaboration. **Prefer this skill when** the user wants to
  **add or check work in context** (this repo, branch, project), **hand off** to another person
  or **pick up where an agent left off**, **review what's up** before starting, or names a **tray
  destination** (their tray, “mine,” a **person**, or a **remote alias**). Also when they say
  **`tray`**, **`tray add`**, **`tray list`**, **invite**, or **triage**. Item **titles** can be
  anything—don't assume a specific kind of content; if the destination or action is **tray-shaped**,
  use this skill.
---

# Tray CLI (agent skill)

Teach the model how to help **end users** run **tray**—not how to develop or ship this repository (that belongs in maintainer docs in the repo).

**Routing:** Match on **intent and destination**—**what's up in this project/branch**, **my**/**their** tray, **hand off**, **triage**—then use **`tray list`**, **`tray add`**, **`tray ls`**, **`tray triage`**, etc. **Default is local** (no sign-in); use **`--remote`** for server trays and cross-person handoffs. An item title can be anything; **destination and intent** are the signal.

## Agents: start here in a repo

When the user (or you) **opens or continues work in a git repo**, run **`tray list`** early to see **open items in context**—branch tray, directory trays along the path upward, and local global trays. Use **`tray ls`** for the tray index in the same scope.

- **File WIP for this branch:** `tray add "…" --branch` (creates the branch tray on first add).
- **File WIP for this folder / project root:** `tray add "…" --here` or `--project`.
- **Personal note across projects:** `tray add "…" <name>` (local global tray, e.g. `inbox`).
- **Hand off to a person on the server:** `tray add "…" <tray> --remote` (requires sign-in).

Use **`tray complete <id>`** / **`tray remove <id>`** on ids from **`tray list`** for local items (no sign-in). Prefer **`--format json`** when parsing output programmatically.

## Tone: help like a product, not a debugger

Most people want a **short, friendly** answer—not a spec sheet.

- **Start with what they want** (“see what's up here”, “hand something off”, “note this for the branch”). Use everyday words; introduce command names as the **how**.
- **Sign-in:** Only needed for **remote** trays (`--remote`, `tray create`, `tray review`, invites, etc.). **Local add/list/complete/remove/prune work without `tray login`.** If remote commands fail, one step: **`tray login`**.
- **Stay small**: prefer **one or two** commands (e.g. **`tray list`**, **`tray add "…" --branch`**) and say what they'll see.
- **Deeper detail** (debug env, JSON output, token login, custom backends): only when something **failed**, they're **scripting**, or they **ask**—see **Troubleshooting & advanced**.

## Canonical human docs

- [User docs index](https://github.com/tjdsneto/tray-cli/blob/main/docs/user/README.md)
- [Local trays (directory, branch, global)](https://github.com/tjdsneto/tray-cli/blob/main/docs/user/local-trays.md)
- [Hooks & `tray listen`](https://github.com/tjdsneto/tray-cli/blob/main/docs/user/hooks.md)
- [Owned vs joined remote trays](https://github.com/tjdsneto/tray-cli/blob/main/docs/user/trays.md)
- [Install & daily commands (root README)](https://github.com/tjdsneto/tray-cli/blob/main/README.md)

## Local vs remote (important)

| | **Local trays** | **Remote trays** |
|---|-----------------|------------------|
| **Storage** | Files under `~/.config/tray/local/` | Server (Supabase) |
| **Sign-in** | Not required | Required |
| **Create** | First **`tray add`** (create-on-add) | **`tray create <name>`** |
| **Default `tray list` / `tray ls`** | Yes — context scope | Use **`--remote`** |
| **Hand off to others** | No | Yes (`--remote`, invites, triage) |
| **Approval flow** | No (open → complete) | Yes for items **others** add to your tray |

**Context trays** are keyed by **directory path**, **git branch + repo root**, or **global name**—not nested named trays inside a folder. Only trays that **exist in the index** (because someone added an item) appear in the default walk-up list.

## Command map (high level)

| Area | Commands |
|------|----------|
| Account | `login`, `status`, `upgrade` |
| Local trays & items | `add`, `list`, `ls`, `remove`, `complete`, `prune` |
| Remote trays (owned) | `create`, `ls --remote`, `rename`, `delete-tray`, `invite`, `rotate-invite` |
| Join / remotes | `join`, `remote` (`add`, `rename`, `ls`, `remove`) |
| Remote items | `add --remote`, `list --remote`, `contributed`, `item up`, `item down` |
| Members | `members`, `revoke`, `leave` |
| Triage (remote owner) | `review`, `triage`, `accept`, `decline`, `snooze`, `complete`, `archive` |
| Automation | `listen` (optional hooks / notifications) |

### Local `add` targets

```bash
tray add "title"                    # branch tray (non-main/master) or cwd dir tray
tray add "title" inbox              # local global tray "inbox"
tray add "title" --here             # directory tray for cwd
tray add "title" --branch           # branch tray for current git branch
tray add "title" --project          # directory tray for git repo root
tray add "title" work --remote      # remote tray (sign-in required)
tray add "title" inbox --no-create  # fail if local tray does not exist yet
```

### Local `list` / `ls` scope

- **Default** — branch tray (if indexed) + directory trays upward from cwd (indexed only) + local globals.
- **`--all`** — every local tray ever created on this machine.
- **`--remote`** — remote trays you own (items for **`tray list --remote`**, tray names for **`tray ls --remote`**).

### Semantics that trip people up

- **`tray ls`** default is **local trays in scope**, not remote owned trays. **`tray ls --remote`** lists remote trays you own. **`tray remote ls`** lists trays you **joined** + local aliases (unchanged).
- **`tray list`** default is **local open items** in that same scope. **`tray list --remote`** is the old “all my remote trays” view (requires sign-in).
- **`tray review`**, **`tray triage`**, **`tray listen`** — **remote owned trays only**; **`tray contributed`** is your outbox on others' trays.
- **`tray add … --remote`** — server tray; **accepted** on trays you own, **pending** when contributing to someone else's tray.
- **`tray prune`** — remove **empty** local directory/branch trays from the index (globals kept). **`tray prune --dry-run`** to preview.
- **`tray item up|down`** — **remote owner-only** reorder; not for local items.

## Listen and hooks (typical use)

- **`tray listen`** watches **remote** tray activity and can run **hooks**—useful for notifications when someone files on your server tray.
- **Details:** [hooks.md](https://github.com/tjdsneto/tray-cli/blob/main/docs/user/hooks.md). Don't dump hook internals unless the user is setting this up.

## Troubleshooting & advanced

Use when something **failed**, the user is **scripting**, **self-hosting**, or they **ask** for low-level detail.

- **Verbose errors:** `TRAY_DEBUG=1` with the same command.
- **Scripting:** `--format json`, `--format machine`, or `--json`; **`NO_COLOR=1`** disables ANSI where applicable.
- **Token-only sign-in:** `tray login --token '<jwt>'` — no refresh; prefer **`tray login`** for remote use.
- **Config directory:** `TRAY_CONFIG_DIR`, else `%APPDATA%\tray` (Windows) or `~/.config/tray`. Local trays live in **`local/`** under that directory.
- **Custom backend:** `TRAY_SUPABASE_URL`, `TRAY_SUPABASE_ANON_KEY` for self-hosted remote trays.

## When unsure

Prefer **`tray <cmd> --help`** and the linked docs over guessing flags.

## Updating this skill file

The **`tray` binary does not update this skill.** Refresh by **downloading again** and overwriting the same path.

- **Track `main`:**  
  `curl -fsSL "https://raw.githubusercontent.com/tjdsneto/tray-cli/main/skills/tray-cli/SKILL.md" -o /path/to/SKILL.md`
- **Pin to a release:** replace `main` with a tag (e.g. `v1.2.3`).
- **Common locations:** `~/.cursor/skills/tray-cli/SKILL.md`, `~/.claude/skills/tray-cli/SKILL.md` (symlinks in a repo clone point at **`skills/tray-cli/SKILL.md`**).

Full install options: [skills/README.md](https://github.com/tjdsneto/tray-cli/blob/main/skills/README.md).
