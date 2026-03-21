# Output formats — brainstorm

This doc is for **ideas and discussion** about how Tray-CLI presents information on stdout/stderr.  
Friendliness should live in **presentation**, not only in translated API errors.

If this file doesn’t show up in `git status` on your machine, check **`.git/info/exclude`** for a local `scratch/` rule and remove it, or run `git add -f scratch/output-formats.md`.

---

## Why this exists

- **REST/API layer** (`http_error.go`, etc.) turns machine responses into short, actionable sentences.
- **Output layer** (tables, hints, success lines, empty states) should feel just as intentional: clear, humane, and helpful without being chatty.

---

## Current behavior (snapshot)

| Surface | Notes |
|--------|--------|
| **Default table** | `NAME`, `ITEMS`, `CREATED` (local TZ); tray list / create |
| **JSON** | Stable fields for scripts (`id`, `item_count`, …) |
| **Markdown** | Tables for pasting into docs/chats |
| **After `create` / `ls` (table)** | “Next steps” block with `tray add`, `tray invite` |
| **Errors** | `tray: …` on stderr; duplicate tray → suggests `tray ls` |
| **Debug** | `TRAY_DEBUG=1` restores raw PostgREST lines |

---

## Principles (draft)

1. **Default output = for humans** — scannable columns, plain language, local time where it helps.
2. **Structured output = for machines** — JSON stays stable, predictable keys; avoid “friendly” strings inside JSON values.
3. **Errors on stderr, success on stdout** — scripts can rely on pipes; exit codes stay meaningful.
4. **One idea per line** — avoid walls of text; use blank lines and short sections when we add more guidance.
5. **Recoverable problems** — whenever we say something failed, pair it with *what to try next* (when it isn’t redundant).

---

## Ideas backlog (not committed — discuss freely)

### Tables

- **Column headers** — Title Case vs `SCREAMING` (current); consider consistency with other CLIs.
- **Empty states** — e.g. “No trays yet.” + one line: how to create the first one (not only on `ls` hints).
- **Widths / alignment** — tabwriter is fine; optional `-o wide` later for IDs or owner info.

### Success & status lines

- **Create success** — brief confirmation line before the table? (“Created tray **work**.”) vs table-only.
- **Idempotent operations** — friendly copy when “already exists” vs strict error (product decision).

### Hints & help-in-output

- **Contextual hints** — only after certain commands (we do this for create/ls); extend to `add` when implemented, etc.
- **Rate / noise** — power users may want `TRAY_QUIET=1` or `--quiet` to suppress hints (future).

### Errors & stderr

- **Visual separation** — optional: stderr prefix `Error:` vs `tray:` only (current).
- **Suggestions** — already partially done; could add `Did you mean …` for typos if we wire cobra suggestions.

### Internationalization

- Out of scope until we have a real i18n plan; keep English copy clear and short for now.

### Accessibility

- Don’t rely on color alone for meaning (if we add ANSI colors later).
- Keep contrast in mind for themes / terminals.

---

## Open questions

- Should **friendly copy** ever appear inside **JSON** (e.g. a `message` field), or keep JSON strictly data?
- Do we want a **`--plain`** flag for no hints, no extra blank lines, stable for logs?
- How much **emoji** (if any) is acceptable in default terminal output? (Currently none.)

---

## References in repo

- Human tray rendering: `internal/output/trays.go`, `tray_hints.go`, `timefmt.go`
- API → user errors: `internal/adapters/postgrest/http_error.go`
- CLI error wrapper: `internal/cli/user_error.go`
- Main stderr prefix: `cmd/tray/main.go`

---

*Add bullets, debate, or paste examples below.*

### Scratch notes

-
