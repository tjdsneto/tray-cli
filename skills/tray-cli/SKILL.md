---
name: tray-cli
description: >-
  Operates the tray CLI (shared inbox / attention queue): Supabase auth, trays,
  items, triage, remotes, and tray listen hooks. Use when the user is using or
  asking about the tray binary, tray login, hooks.json, TRAY_* env vars, or
  automation around this CLI.
---

# Tray CLI (agent skill)

Teach the model how to help **end users** of the **`tray`** binary—not how to hack this repository (that lives in maintainer docs and `CLAUDE.md` / `.cursor/rules`).

## Tone: help like a product, not a debugger

Most people want a **short, friendly** answer—not a spec sheet.

- **Casual asks** (“check my tray”, “what’s on my tray”, “show my inbox”): Use **plain language** first. If they’re **not signed in**, say that in one sentence and give **one** next step: run **`tray login`** (browser sign-in, usually Google). Do **not** open with **exit codes**, **`tray status` exit 1**, **`--force`**, **`--token` / JWT**, or **`TRAY_DEBUG`** unless they are **debugging a failure** or **explicitly ask** for advanced options.
- **After they’re signed in**, keep it minimal: e.g. **`tray list`** for items on trays they own, **`tray ls`** for tray names—**briefly say what each shows** instead of dumping every related subcommand at once.
- **Reserve** technical detail (exit codes, token-only login caveats, verbose debug env) for **troubleshooting** or **power users**—see **Session** below when relevant.

## Canonical human docs

- [User docs index](https://github.com/tjdsneto/tray-cli/blob/main/docs/user/README.md)
- [Hooks & `tray listen`](https://github.com/tjdsneto/tray-cli/blob/main/docs/user/hooks.md)
- [Owned vs joined trays, list semantics](https://github.com/tjdsneto/tray-cli/blob/main/docs/user/trays.md)
- [Install & daily commands (root README)](https://github.com/tjdsneto/tray-cli/blob/main/README.md)

## Config and backend

- **Config directory:** `TRAY_CONFIG_DIR`, else Windows `%APPDATA%\tray`, else `$XDG_CONFIG_HOME/tray` or `~/.config/tray`.
- **Supabase:** `TRAY_SUPABASE_URL`, `TRAY_SUPABASE_ANON_KEY` (env overrides build-time embeds). See `.env.example` in the repo when helping someone clone or self-host the client.

## Session

- **`tray login`** — OAuth in browser (local page is **Google** only; use **`--provider`** for other IdPs); stored refresh token; CLI refreshes JWT when needed. **`tray login --force`** to re-prompt even if a session exists.
- **`tray login --token '<jwt>'`** — manual access token only; **no refresh**; prefer OAuth for long-term use.
- **`tray status`** — verify credentials; **`--format json`** for scripts (exit **0** if signed in, **1** if not). **Do not quote exit codes** to casual users—just say “not signed in” and point to **`tray login`**.

## Command map (high level)

| Area | Commands |
|------|----------|
| Account | `login`, `status`, `upgrade` |
| Trays (owned) | `create`, `ls`, `rename`, `delete-tray`, `invite`, `rotate-invite` |
| Join / remotes | `join`, `remote` (`add`, `rename`, `ls`, `remove`) |
| Items | `add`, `list`, `remove`, `contributed`, `item up`, `item down` |
| Members | `members`, `revoke`, `leave` |
| Triage (owner) | `review`, `triage`, `accept`, `decline`, `snooze`, `complete`, `archive` |
| Automation | `listen` (with `hooks.json`) |

Semantics that trip people up:

- **`tray ls`** — trays **you own**. **`tray remote ls`** — trays you **joined** + local aliases.
- **`tray list`** — items on **your** trays; **`tray list <tray>`** only for trays **you own**.
- **`tray contributed`** — items **you** filed on **others’** trays (outbox).
- **`tray add "title" <tray>`** — target tray by **name**, **id**, or **remote alias**; new items are **accepted** on trays **you own**, **pending** when you **contribute** to someone else’s tray.
- **`tray item up|down <item-id>`** — **owner-only**: swap manual list order (`sort_order`) with the neighbor above or below; lists and triage use this order by default (see `#` / `sort_order` in `tray list --format json`).

## Output formats

Stable scripting: **`--format json`**, **`--format machine`**, or **`--json`**. Human default: tables and hints. **`NO_COLOR=1`** disables ANSI where applicable.

## Listen and hooks

- **`tray listen`** polls/subscribes and runs **`hooks.json`** (default `$TRAY_CONFIG_DIR/hooks.json`; override with `--hooks`, or **`--no-hooks`**).
- Events such as **`item.pending`**, **`item.accepted`**, **`item.declined`**, **`item.completed`**; hook processes receive **`TRAY_*`** environment variables—**full list and recipes** belong in [hooks.md](https://github.com/tjdsneto/tray-cli/blob/main/docs/user/hooks.md).
- **`TRAY_DEBUG=1`** — verbose errors (e.g. PostgREST bodies) when something fails; use when diagnosing API issues.

## When unsure

Prefer **`tray <cmd> --help`** and the linked docs over guessing flags. Do not invent Supabase dashboard steps beyond “enable provider / set redirect URI” already described in the README.

## Updating this skill file

The **`tray` binary does not update this skill.** If the user installed **`SKILL.md`** with `curl` (or copied it by hand), they refresh it by **downloading again** and overwriting the same path.

- **Track `main` (newest doc changes):**  
  `curl -fsSL "https://raw.githubusercontent.com/tjdsneto/tray-cli/main/skills/tray-cli/SKILL.md" -o /path/to/SKILL.md`
- **Pin to a release:** replace `main` with a tag (e.g. `v1.2.3`) in that URL; bump the tag when they want a newer snapshot.
- **Common locations:** `~/.cursor/skills/tray-cli/SKILL.md`, `~/.claude/skills/tray-cli/SKILL.md` (use the path that applies; if one is a symlink, overwriting the real file is enough).

Full install options and symlink notes: [skills/README.md](https://github.com/tjdsneto/tray-cli/blob/main/skills/README.md).
