# Trays: yours, joined, and aliases

## Two kinds of tray lists

| Command | What it shows |
|--------|----------------|
| **`tray ls`** | Trays **you own** (you created them). This is your primary inbox list. |
| **`tray remote ls`** | Trays **you joined** as a member (someone else’s tray), plus any **local aliases** you saved. Joined trays appear here even if you never set a nickname. |

Server-side tray **names** (e.g. `work`) are chosen by the **owner**. Everyone who can see that tray sees the **same name**. A nickname like `tiago-work` exists only as a **local alias** on your machine (`tray join … <alias>` or `tray remote add`).

**Joined members** can **add** lines to someone else’s tray (invite / `tray join`), but they **do not** get a full read of that tray’s inbox — only the **owner** triages everyone’s items. As a contributor you still see **your own** lines (e.g. via **`tray contributed`** and when you add or follow up on something you filed).

## Items

| Command | What it shows |
|--------|----------------|
| **`tray list`** (no args) | Items on trays **you own** only. Human output is **grouped by status** (Accepted → Pending → … → Completed); section titles use **status colors** when the terminal supports it. Each item is a **summary line** (`order · tray · from · when`) plus the **full title** wrapped to the terminal width. The leading **order** is manual within that tray (`tray item up` / `down`). **Completed** rows show completion time in the summary line where other sections show when the item was added. |
| **`tray list <tray>`** | Items on that tray only if **you own** it. The name/id/alias is resolved **among your owned trays only** — a joined tray with the same name as yours does not collide. |
| **`tray review`**, **`tray triage`** | Pending items on **your** trays only (owner triage). Same tray rules as **`tray list`**. |
| **`tray listen`** | Pending snapshot/poll targets **owned** trays only (aligned with **`tray review`**). Outbox hooks are separate. |
| **`tray contributed`** | Items **you** added to **someone else’s** trays (your “outbox”). |

To add an item to a tray you **don’t** own, use the tray’s **name** (from `tray remote ls`), a **remote alias**, or the tray id — see **`tray add --help`**.

## See also

- **[Listen hooks](hooks.md)** — `tray listen`, `hooks.json`, `TRAY_*` env vars.
