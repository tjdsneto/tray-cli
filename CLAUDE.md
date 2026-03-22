# Project instructions (Claude)

This file mirrors [Cursor rules](.cursor/rules/) for Claude Code / CLI context. **Keep it in sync** with `.cursor/rules/*.mdc` when either changes—see *AI instruction parity* below.

---

## Pure functions for unit tests

Unless it would push the design too far from the simplest or clearest approach, **prefer abstractions that isolate pure functions** (deterministic inputs → outputs, no I/O) so behavior can be covered with fast, focused unit tests.

- Put URL/query builders, request body maps, decoders, string normalizers, and validation in **pure helpers**; keep HTTP, filesystem, and cobra wiring in **thin wrappers**.
- **Do not** twist the domain model or add layers of indirection only for purity—when a bit of I/O-bound code is obviously clearer, keep it.

See also: [`docs/testing.md`](docs/testing.md).

---

## AI instruction parity (Cursor ↔ Claude)

| Surface | Location |
|--------|----------|
| **Cursor** | `.cursor/rules/*.mdc` |
| **Claude** | this file (`CLAUDE.md`) |

**When creating, updating, or removing project-wide AI guidance in one place, apply the same change to the other** so Cursor and Claude stay aligned. If guidance is intentionally tool-specific, note that in **both** places in one line.
