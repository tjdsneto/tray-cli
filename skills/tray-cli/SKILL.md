---
name: tray-cli
description: >-
  Helps people use the **tray** CLI for shared **inboxes** (**trays**), **items**,
  **handoffs**, **triage**, and **attention** on queued work. **Prefer this skill when** the user
  wants to **add something to their own tray or someone else’s**, **hand off** work to another
  person, **check or review** what’s on a tray, or names a **tray destination** (their tray,
  “mine,” a **person’s name**, or a **remote alias**). Also when they say **`tray`**, **`tray add`**,
  **invite**, or **triage**. Item **titles** can be anything—don’t assume a specific kind of
  content (links, reviews, todos, etc.); if the destination or action is **tray-shaped**, use this skill.
---

# Tray CLI (agent skill)

Teach the model how to help **end users** run **tray**—not how to develop or ship this repository (that belongs in maintainer docs in the repo).

**Routing:** Match on **intent and destination**—**my**/**their**/**someone’s tray**, a **person or alias**, **hand off**, **what’s on my tray**, **triage**—then use **`tray add`**, **`tray list`**, **`tray triage`**, **`tray review`**, etc., resolving targets with **`tray remote ls`** when needed. An item title can be anything; **don’t** treat particular phrasing or formats as the signal—**destination and intent** are.

## Tone: help like a product, not a debugger

Most people want a **short, friendly** answer—not a spec sheet.

- **Start with what they want** (“see my inbox”, “hand something off”, “clean up my queue”). Use everyday words; introduce command names as the **how**, not the opening jargon.
- **Sign-in**: If they’re not signed in or nothing works, one sentence and **one** step: run **`tray login`** (browser sign-in). Don’t lead with flags, env vars, or raw error formats.
- **Stay small**: prefer **one or two** commands that answer the question (e.g. **`tray list`**, **`tray ls`**, **`tray triage`**) and say what they’ll see—don’t dump the whole CLI.
- **Deeper technical detail** (debug env vars, machine-readable output, token login, custom backends): only when something **failed**, they’re **scripting**, they **self-host**, or they **explicitly ask**—see **Troubleshooting & advanced** below.

## Canonical human docs

- [User docs index](https://github.com/tjdsneto/tray-cli/blob/main/docs/user/README.md)
- [Hooks & `tray listen`](https://github.com/tjdsneto/tray-cli/blob/main/docs/user/hooks.md)
- [Owned vs joined trays, list semantics](https://github.com/tjdsneto/tray-cli/blob/main/docs/user/trays.md)
- [Install & daily commands (root README)](https://github.com/tjdsneto/tray-cli/blob/main/README.md)

## Sign-in and status (typical use)

- **`tray login`** — Sign in in the browser (often Google on the first screen). **`tray login --force`** if they need to sign in again or switch accounts.
- **`tray status`** — Check whether the CLI is signed in. Answer in plain language (“you’re signed in” / “you need to sign in”); don’t lead with exit codes or JSON unless they’re scripting or debugging.

## Command map (high level)

| Area | Commands |
|------|----------|
| Account | `login`, `status`, `upgrade` |
| Trays (owned) | `create`, `ls`, `rename`, `delete-tray`, `invite`, `rotate-invite` |
| Join / remotes | `join`, `remote` (`add`, `rename`, `ls`, `remove`) |
| Items | `add`, `list`, `remove`, `contributed`, `item up`, `item down` |
| Members | `members`, `revoke`, `leave` |
| Triage (owner) | `review`, `triage`, `accept`, `decline`, `snooze`, `complete`, `archive` |
| Automation | `listen` (optional hooks / notifications) |

Semantics that trip people up:

- **`tray ls`** — trays **you own**. **`tray remote ls`** — trays you **joined** + local aliases.
- **`tray list`** — items on **your** trays; **`tray list <tray>`** only for trays **you own** (names resolve among **owned** trays only, so a joined tray with the same name does not collide).
- **`tray review`**, **`tray triage`**, **`tray listen`** — **owned trays only** for your inbox workflow (joined trays don’t work like your own tray here); **`tray contributed`** is the outbox of lines **you** filed elsewhere.
- **`tray contributed`** — items **you** filed on **others’** trays (outbox). Joining someone’s tray lets you **add** lines there; it does **not** give you their full inbox — only the **owner** sees the full queue on that tray.
- **`tray add "title" <tray>`** — target tray by **name**, **id**, or **remote alias**; new items are **accepted** on trays **you own**, **pending** when you **contribute** to someone else’s tray.
- **`tray item up|down <item-id>`** — **owner-only**: move an item up or down in the list order on that tray.

## Listen and hooks (typical use)

- **`tray listen`** watches for tray activity and can run **hooks** (small scripts) you configure—useful for notifications or glueing tray into other tools.
- **Details** (config file location, event names, environment variables hooks receive): [hooks.md](https://github.com/tjdsneto/tray-cli/blob/main/docs/user/hooks.md). Don’t dump hook internals unless the user is setting this up or troubleshooting listen.

## Troubleshooting & advanced

Use this block when something **failed**, the user is **scripting**, **self-hosting**, or they **ask** for low-level detail—not for first-time “how do I use tray?” questions.

- **Verbose errors:** `TRAY_DEBUG=1` with the same command can surface more detail when diagnosing a failure (see maintainer docs if they’re hacking on the client).
- **Scripting / stable output:** `--format json`, `--format machine`, or `--json`; **`NO_COLOR=1`** disables ANSI where applicable.
- **Token-only sign-in:** `tray login --token '<jwt>'` — manual token, **no refresh**; prefer normal **`tray login`** for day-to-day use.
- **Config directory:** `TRAY_CONFIG_DIR`, else Windows `%APPDATA%\tray`, else `$XDG_CONFIG_HOME/tray` or `~/.config/tray`.
- **Custom / self-hosted backend:** `TRAY_SUPABASE_URL`, `TRAY_SUPABASE_ANON_KEY` (override built-in defaults). Repo **`.env.example`** and maintainer docs apply—don’t improvise dashboard steps beyond what the README already says for providers and redirect URIs.

## When unsure

Prefer **`tray <cmd> --help`** and the linked docs over guessing flags.

## Updating this skill file

The **`tray` binary does not update this skill.** If the user installed **`SKILL.md`** with `curl` (or copied it by hand), they refresh it by **downloading again** and overwriting the same path.

- **Track `main` (newest doc changes):**  
  `curl -fsSL "https://raw.githubusercontent.com/tjdsneto/tray-cli/main/skills/tray-cli/SKILL.md" -o /path/to/SKILL.md`
- **Pin to a release:** replace `main` with a tag (e.g. `v1.2.3`) in that URL; bump the tag when they want a newer snapshot.
- **Common locations:** `~/.cursor/skills/tray-cli/SKILL.md`, `~/.claude/skills/tray-cli/SKILL.md` (use the path that applies; if one is a symlink, overwriting the real file is enough).

Full install options and symlink notes: [skills/README.md](https://github.com/tjdsneto/tray-cli/blob/main/skills/README.md).
