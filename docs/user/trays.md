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
| **`tray list`** (no args) | Items on trays **you own** only. |
| **`tray list <tray>`** | Items on that tray only if **you own** it. |
| **`tray contributed`** | Items **you** added to **someone else’s** trays (your “outbox”). |

To add an item to a tray you **don’t** own, use the tray’s **name** (from `tray remote ls`), a **remote alias**, or the tray id — see **`tray add --help`**.

## See also

- **[Listen hooks](hooks.md)** — `tray listen`, `hooks.json`, `TRAY_*` env vars.
