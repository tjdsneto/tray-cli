# Tray-CLI

**Tray** is a **command-line tool** for shared **inbox trays**: add items to your tray or someone else’s, **triage** what lands on yours, and keep **asks and handoffs** from dissolving into chat scrollback.

## Why not just DM or use chat?

**Chat and DMs** work in the moment—but “I’ll get to that” is easy to **forget**, and threads get **buried** under whatever came next. Tray gives each owner a **persistent queue**: items stay on **your tray** until **you** act on them; people you **invite** can still file requests without a separate project tool.

Because it runs in your **terminal**, **AI coding assistants** can use it alongside you in the same place you already work, and you can **file or check a handoff** without **switching** to another app or chat window just to message someone. Optional **[agent skills](#agent-skills)** ([**`skills/README.md`**](skills/README.md)) teach assistants how to help with `tray`.

**Get started:** [Install](#install) → [First use](#first-use) → [Triage](#triage-your-inbox) → optional **[`tray listen`](docs/user/hooks.md)** for notifications.

**More documentation:** **[`docs/user/`](docs/user/README.md)** · [`docs/README.md`](docs/README.md) · **[`skills/`](skills/README.md)**.

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/tjdsneto/tray-cli/main/scripts/install.sh | bash
```

No **`git clone`** required.

## First use

```bash
tray login
tray create my-tray
tray invite my-tray
```

Complete sign-in when your browser opens. **`tray status`** shows whether you’re signed in; **`tray login --force`** runs sign-in again (for example if you switch accounts).

**Invite others:** **`tray invite`** prints a token or URL. Share it; teammates run **`tray join <token-or-url>`** (optional local alias: **`tray join … nickname`**). See **[`docs/user/trays.md`](docs/user/trays.md)** for owned vs joined trays.

**Add your own items:** use your tray like a personal inbox—file things for yourself or leave reminders:

```bash
tray add "Your task title" my-tray
```

**See what’s on your tray:** **`tray list`** shows items on trays **you own** (all of them, or focus with **`tray list my-tray`**). You’ll see lines **you** added and anything **others** filed if they joined—one place to review the whole queue.

## Upgrades

**`tray upgrade`**, or run the [install](#install) `curl` line again for the latest release. Pin a version: **`tray upgrade --version v0.1.x`** or **`TRAY_VERSION=v0.1.x`** before the `curl` command.

## Agent skills

Skills teach assistants how to help with **`tray`**. **Claude Code** can install the **`tray-cli`** plugin from this repo’s marketplace (bundles **`skills/tray-cli/SKILL.md`**) or you can **`curl`** the skill file alone. **Cursor** has no plugin marketplace for skills—use **`curl`**. More detail and edge cases: **[`skills/README.md`](skills/README.md)**.

**Cursor** — personal skill (`~/.cursor/skills/tray-cli/SKILL.md`):

```bash
mkdir -p ~/.cursor/skills/tray-cli
curl -fsSL "https://raw.githubusercontent.com/tjdsneto/tray-cli/main/skills/tray-cli/SKILL.md" \
  -o ~/.cursor/skills/tray-cli/SKILL.md
```

**Claude Code** — pick one:

**Plugin marketplace** (recommended; refresh after updates on GitHub):

```bash
claude plugin marketplace add tjdsneto/tray-cli
claude plugin install tray-cli@tjdsneto
claude plugin marketplace update   # re-run when you want the latest from main
```

**Working inside a clone of this repo?** From the repository root:

```bash
claude plugin marketplace add .
claude plugin install tray-cli@tjdsneto
```

**Manual `curl`** (only the skill file, `~/.claude/skills/tray-cli/SKILL.md`):

```bash
mkdir -p ~/.claude/skills/tray-cli
curl -fsSL "https://raw.githubusercontent.com/tjdsneto/tray-cli/main/skills/tray-cli/SKILL.md" \
  -o ~/.claude/skills/tray-cli/SKILL.md
```

For **`curl`**, use a **release tag** instead of **`main`** in the URL to match a specific [release](https://github.com/tjdsneto/tray-cli/releases). Re-run the same **`curl`** to refresh.

## Trays

**`tray create <name>`** creates a tray; **`rename <tray> <new-name>`** (owner); **`delete-tray`** (owner); **`ls`** lists **only trays you own**; **`join <token-or-url> [local-alias]`** joins via invite; **`invite`** / **`rotate-invite`** manage invite tokens (owner); **`members`**, **`revoke`**, **`leave`** for membership.

**Joined trays** (someone else’s inbox you were invited to) show up in **`tray remote ls`**, with optional local aliases — see **[`docs/user/trays.md`](docs/user/trays.md)**.

## Items

**`tray add "title" <tray>`** adds a pending item (tray = name, id, or **`remote`** alias). **`list`** shows items on **your** trays; **`list <tray>`** requires an **owned** tray. **`contributed`** lists items you filed on others’ trays; **`remove <item-id>`** (owner deletes any item; contributor can delete own **pending** items).

## Triage your inbox

If you **own** a tray, items from teammates (and you) show up for you to handle. Scan what’s pending with **`tray review`** or the full-screen **`tray triage`** UI, then move work forward: **`accept`**, **`decline`**, **`snooze`**, **`complete`**, or **`archive`**. Item ids for scripts come from **`tray list --format json`**. Flags and examples: **`tray <command> --help`** and **[`docs/user/`](docs/user/README.md)**.

## Listen hooks

**`tray listen`** can run **`hooks.json`** when items change (pending on your trays, outbox accepted/declined/completed). See **[`docs/user/hooks.md`](docs/user/hooks.md)** for events, environment variables, and recipes.

## Remote aliases

**`join … <alias>`** or **`tray remote add <alias> <invite-url-or-token>`** saves `remotes.json`. **`tray remote rename <current> <new>`** renames an alias. **`tray remote ls`** / **`tray remote remove <alias>`** manage that file.

## Output (list-style commands)

**Default is human-friendly:** tables, local dates, and contextual hints where we’ve added them.

| Flag | Purpose |
|------|---------|
| **`--format human`** (default) | Friendly tables and hints |
| **`--format json`**, **`--format machine`**, or **`--json`** | Stable JSON for scripts |
| **`--format markdown`** / **`md`** | Markdown tables |

`--json` is shorthand for `--format json` and must not be combined with another explicit format.

For **trays**, the default **human** output shows **name**, **item count**, and **created** (local timezone; set **`TZ`** if needed). Tray **IDs** and **`item_count`** appear in **`--format json`**.

For **items** (`list`, `review`, …), human output includes **who added** the item (`you` vs a short id), **created** as a **relative time** when recent, and **status** colors on a TTY. **`tray triage`** is an interactive pending queue (TTY); **`tray review`** is non-interactive.

## Install troubleshooting

**`tray: command not found`** — the installer usually prints a line to add its directory to your **`PATH`** (common when it used **`~/.local/bin`**). Run that, or open a new terminal.

**Permissions / `sudo`** — the script does not use **`sudo`** unless you set **`TRAY_INSTALL_USE_SUDO=1`**. It picks a writable directory; override with **`TRAY_INSTALL_DIR`**.

**Install with Go** (optional): **`go install github.com/tjdsneto/tray-cli/cmd/tray@latest`** (Go **1.25+**; add **`$(go env GOPATH)/bin`** to **`PATH`**).

---

## Maintainers and contributors

Everything for **developing this repo**—tests, local builds, releases, migrations, OAuth internals, **`TRAY_DEBUG`**, install script details, and architecture—is in **[`docs/maintainers/README.md`](docs/maintainers/README.md)**. Start there if you are cloning the source, running **`./run.sh`**, or cutting a release.
