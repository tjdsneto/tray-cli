# macOS menu bar (`tray bar`)

**`tray bar`** puts a small **menu bar icon** on macOS so you can see how many tray items need attention, browse them by tray, and **copy a handoff line** for an agent session. It is separate from **`tray listen`** (hooks/notifications)—use both if you want.

**Platform:** macOS only. On Linux or Windows, `tray bar` exits with a friendly “macOS only” message.

## Start it

```bash
tray bar
```

Runs in the **foreground** (the menu bar icon lives in this process). Quit from the menu or stop the terminal process.

**Background (daemon):**

```bash
tray bar --daemon
```

Writes **`bar.pid`** and **`bar.log`** under your config directory (`$TRAY_CONFIG_DIR`, default `~/.config/tray`). Only one daemon at a time; a second start reports “already running.”

**Refresh interval:**

```bash
tray bar --interval 1m
```

Default **`30s`**. The menu also has **Refresh**; remote/local data is polled on each tick.

## What the badge counts

The number on the icon = **local open items** + **remote pending items on trays you own**.

- **Signed out:** local open only.
- **Signed in:** adds remote pending on your trays.
- **Zero:** icon shows **`tray`** with no number (no `0` badge).
- If remote fetch fails, the bar keeps the **last good snapshot** and may show a muted **“remote unavailable”** line in the menu.

## Menu

Click the icon to open a dropdown of **attention backlog only** (same set as the badge):

- **Grouped by tray name** (A–Z). Local and remote trays share grouping; if the same name exists in both scopes, sections are labeled **`inbox · local`** / **`inbox · remote`**.
- Within a tray, **newest first**.
- **Footer:** **Refresh**, **Quit**.

Click a row to copy exactly two tokens to the clipboard:

```text
<tray-ref> <item-id>
```

Example: `inbox a1b2c3d4`. If the tray-ref has spaces, it is double-quoted (e.g. `"dir:/Users/me/My Projects" a1b2c3d4`). No title, JSON, or pre-built command—agents resolve details with the CLI.

## Accept and the badge

When **you** (or an agent) **`tray accept`** a **remote pending** item, it **leaves the menu** on the next refresh and the **badge drops**—even before **`tray complete`**. Accepted-but-in-progress work is not shown in the bar.

Local open items stay until **`tray complete`** (or **`tray remove`**).

## Hand off to an agent

Typical flow:

1. Click a menu row → paste **`tray-ref item-id`** into your agent chat.
2. Agent **fetches** item details (`tray list`, `tray review`, or `--format json` as needed).
3. If the item is **remote pending**, run **`tray accept <id>`** before doing the work.
4. Do the work, then **`tray complete <id>`**.

**Local open:** skip accept. **Already accepted:** skip accept; complete when done.

The **[`tray-cli` skill](../../skills/tray-cli/SKILL.md)** documents this handoff for AI assistants.

## See also

- **[Listen hooks](hooks.md)** — background notifications via `hooks.json` (remote owned trays).
- **[Local trays](local-trays.md)** — directory, branch, and global trays on this machine.
- **[Owned vs joined remote trays](trays.md)** — pending vs open semantics.

CLI flags: **`tray bar --help`**.
