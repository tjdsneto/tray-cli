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
| **`tray list`** (no args) | Items on trays **you own** only. Human output is **grouped by status** (Accepted → Pending → … → Completed), then **by tray name** under each status. Section titles use **status colors** when the terminal supports it; tray names are a dim subheading. Each item is one **meta line** (`<display #> · from · when · full item id`; the **tray name** is only in the subheading above that group). **`#`** is a **1-based line number** within that tray block in the listing (for readability). Rows are still **sorted by manual tray order** (`sort_order`, adjusted with `tray item up` / `down`). On a color TTY, the **tray name**, **line number**, and **who · when** are dim; the **uuid** and **title** use default brightness so the description and id are easiest to read. If the terminal is narrow, the middle of the meta line may ellipsize so the **uuid stays visible**—then **indented title** lines only. **Completed** rows show completion time in the meta line where other sections show when the item was added. Commands that take `<item-id>` also accept a **unique hex prefix** of at least **8** characters (hyphens optional). The prefix is matched only in a **small command-specific set** (for example **accept** and **decline** only among **pending** lines on trays you own, like **`tray review`**; **complete** and **snooze** among **pending and accepted** on your trays so a prefix still works after you accept something; **archive** among **pending, accepted, and snoozed**; **item up/down** among all items on trays you own; **remove** among your trays plus your **contributed** lines). If the prefix matches more than one item in that set, the command fails and asks for a longer prefix or the full id. |
| **`tray list <tray>`** | Items on that tray only if **you own** it. The name/id/alias is resolved **among your owned trays only** — a joined tray with the same name as yours does not collide. |
| **`tray review`**, **`tray triage`** | Pending items on **your** trays only (owner triage). Same tray rules as **`tray list`**. |
| **`tray listen`** | Pending snapshot/poll targets **owned** trays only (aligned with **`tray review`**). Outbox hooks are separate. |
| **`tray contributed`** | Items **you** added to **someone else’s** trays (your “outbox”). |

To add an item to a tray you **don’t** own, use the tray’s **name** (from `tray remote ls`), a **remote alias**, or the tray id — see **`tray add --help`**.

## See also

- **[Listen hooks](hooks.md)** — `tray listen`, `hooks.json`, `TRAY_*` env vars.
