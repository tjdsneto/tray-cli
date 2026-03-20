# Tray-CLI — final specs (decisions)

Single place for choices locked after brainstorming. Deeper rationale and checklists live in the linked scratch docs.

---

## Product and repository

| Item | Decision |
|------|----------|
| **Product name** | **Tray-CLI** (display/marketing; inbox-tray product, CLI-first). |
| **GitHub repository** | **`tray-cli`** — **public**. |
| **CLI binary on `$PATH`** | **`tray`** (short command; see [cli-design.md](cli-design.md)). Release artifacts may be named e.g. `tray-darwin-arm64` to match OS/arch. |

Historical folder name **attention-queue** and older doc references may still appear locally until the repo is renamed or migrated; treat **`tray-cli`** as the canonical remote name going forward.

---

## Distribution

| Item | Decision |
|------|----------|
| **Binaries** | Published via **GitHub Releases** (per OS/arch, e.g. darwin arm64/amd64, linux amd64, windows amd64). |
| **Install** | **Install script** (e.g. `curl … \| sh`) that **detects OS/arch**, downloads the matching release asset, installs the binary into a directory on **`$PATH`** (or a documented default such as `~/bin` + instructions). |

Details and alternatives: [research-and-stack.md](research-and-stack.md) (CLI distribution), [repo-visibility-and-distribution.md](repo-visibility-and-distribution.md).

---

## Backend and portability

| Item | Decision |
|------|----------|
| **Backend (v1)** | **Supabase** — Postgres, Auth, Row Level Security, PostgREST (Data API), optional Edge Functions / Realtime later. |
| **Code structure** | **Abstract the backend behind a small internal boundary** (e.g. an interface for “auth session”, “trays/buckets”, “items”, “join/invite”) so the CLI and domain logic do not call Supabase/PostgREST types directly everywhere. Implementations: **Supabase/HTTP first**; a future **Firebase** (or other) adapter would be a **deliberate port** of that interface—not a rewrite of the whole CLI, but still real work (different auth, rules model, query shapes). |

Goal: switching providers is **possible without rewriting every command**, not that it is ever zero-cost.

---

## Stack summary

| Layer | Choice |
|-------|--------|
| CLI | **Go** (e.g. cobra); commands and flags per [cli-design.md](cli-design.md). |
| Backend | **Supabase** (abstracted in code as above). |
| Repo visibility | **Public** |
| Releases + install | **GitHub Releases** + **OS/arch install script** → `$PATH` |

---

## References

| Doc | Role |
|-----|------|
| [PRODUCT.md](PRODUCT.md) | Product framing |
| [cli-design.md](cli-design.md) | Commands, config dir, `--json` |
| [backend-spec.md](backend-spec.md) | API shape, tables, RLS |
| [launch-requirements.md](launch-requirements.md) | MVP checklist |
| [research-and-stack.md](research-and-stack.md) | Stack research, distribution notes |

---

*Created: 2026-03-20. Update this file when decisions change.*
