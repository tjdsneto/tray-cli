# Item body (larger payloads) — design

**Date:** 2026-10-22  
**Status:** Approved for planning  
**Related:** chatgpt-handoff skill (title-only packets + ~12k soft cap); agent session trays; backlog `todos.md` 2026-10-22

## Problem

Tray items are **title-only**. ChatGPT handoff and similar flows encode everything in `title`, with a ~12k-character soft cap in the skill. Full artifacts (e.g. a 20k-character document) cannot fit one item without condensing or splitting. Callers need a first-class place for large text without bloating list headlines.

## Goals

- Optional **`body`** text field on items (remote + local), set at create time.
- Hard cap **100_000 Unicode runes** enforced in CLI and gateway (clear errors).
- Full stack: CLI, PostgREST/domain, gateway, local store, user docs, and handoff skills.
- Keep **title** as the short headline / packet headline for lists and triage UI.
- Structured/`--json` output includes `body` when present so agents can read full content without a new command.

## Non-goals (v1)

- Editing or patching `body` after create.
- Attachments, blobs, or object storage.
- Separate `item_bodies` table or lean column subsets for list.
- Raising the chatgpt-handoff **title** soft cap (long content moves to `body`).
- New `tray show` command.
- Streaming / chunked upload.

## Approach

**Nullable `body text` on `items`** (Approach 1).

| Concept | Choice |
|--------|--------|
| Storage | `items.body text` nullable; empty/whitespace → null/omitted |
| Cap | 100_000 runes; CLI + gateway validate; optional Postgres `check` for defense in depth |
| Write | Create only via `tray add` / gateway create |
| Read (human) | Table list / triage / menu bar: title only; optional compact “has body” marker |
| Read (agent) | JSON / structured list includes `body` when set |
| Skills | Short packet stays in title; full documents in body |

Rejected for v1: raise title-only soft cap (Approach A); separate bodies table (Approach 2); `payload jsonb` (Approach 3).

## Architecture

```
Writer                                      Store                         Reader
──────                                      ─────                         ──────
tray add "Headline" --body-file doc.md
  or gateway POST { title, body }
        │
        ├── validate ≤ 100k runes
        ├── insert items (title + body)
        │
        └── local JSON / PostgREST / gateway
                                              list (table): title [·]
                                              list --json: { title, body, … }
                                              chatgpt-handoff: title=packet, body=artifact
```

### Layers

1. **Migration** — `alter table public.items add column body text;` (+ optional length check).
2. **Domain** — `Item.Body *string` (or empty string = absent); `ItemService.Add` takes optional body.
3. **PostgREST adapter** — include `body` in insert map, select columns, row decode.
4. **Gateway** — accept optional `body` on create; same validation + insert path.
5. **Local store** — `body` on item JSON when non-empty.
6. **CLI** — `--body` / `--body-file` (mutually exclusive) on `tray add` for local, `--remote`, and `--agent-session-id`.
7. **Output** — table omits body text; JSON includes it; optional marker when body present.
8. **Docs / skills** — user docs + chatgpt-handoff / tray-cli: title vs body convention.

### Invariants

- `title` remains required and non-empty after trim.
- Body is optional; trim-empty means “no body”.
- Over-cap rejects before network when possible (CLI) and at gateway edge.
- RLS unchanged — `body` is a normal column on `items`.
- Delete item deletes body with the row.

### Errors

| Condition | Behavior |
|-----------|----------|
| Body > 100_000 runes | Clear error: body exceeds limit |
| Both `--body` and `--body-file` | Flag error |
| `--body-file` missing/unreadable | Normal file error |
| Empty body after trim | Omit field (same as not passing) |

## Testing

Prefer pure helpers:

- Rune-cap validator (under/at/over limit).
- Add-request body map: include vs omit `body`.
- Local store round-trip with body.
- Gateway mutation/handler: optional body + oversize reject.
- CLI mutual-exclusion of `--body` / `--body-file` where cheap.

## Success criteria

- A 20k+ character document can be attached as `body` on one item while title stays a short headline.
- Remote and local add paths both accept body; gateway create does too.
- Agents can retrieve full body via existing list JSON (or equivalent structured output).
- chatgpt-handoff no longer requires truncating full artifacts into the title.
