# Backend Specification

API, auth, multi-tenancy, and **database structure** for **Tray CLI** (repo: attention-queue). The **Tray CLI** (Go) and any future client talk to this backend over REST. Local client state (remotes, credentials, config) is described in [cli-design.md](cli-design.md#local-storage-local-db).

---

## 1. Overview

- **Stack (MVP):** **Supabase** — PostgreSQL, Auth, RLS, optional Realtime. Optional Edge Functions for invite/join if not covered by PostgREST.
- **API style:** REST. Tray CLI uses REST (or Supabase client); same API can serve a web UI or MCP server.
- **Multi-tenant:** One “tenant” = one user’s trays (their buckets). Data isolated by `owner_id` and bucket membership.

---

## 2. Auth

- **Provider:** Supabase Auth (email/password, OAuth, magic links).
- **CLI flow:** User runs `tray login`; browser or device flow completes; backend returns session (access token + refresh token). CLI stores tokens in local config (see [cli-design](cli-design.md#local-storage-local-db)).
- **Requests:** CLI sends `Authorization: Bearer <access_token>`. Backend (Supabase) validates JWT and sets `auth.uid()` for RLS.
- **Refresh:** Use refresh token to obtain new access token when expired; CLI handles this transparently.

---

## 3. API endpoints (contract for CLI)

The CLI needs at least the following. Exact paths and request/response shapes to be defined (OpenAPI later); this is the logical contract.

| Area | Method / action | Purpose |
|------|-----------------|---------|
| **Auth** | Login (browser/device flow) | Return session (access + refresh token). |
| **Buckets** | `POST /buckets` | Create bucket (name, owner_id from auth). |
| **Buckets** | `GET /buckets` | List buckets where I am owner. |
| **Buckets** | `GET /buckets/:id/invite` or `POST /buckets/:id/invite` | Get or generate shareable invite (URL/token). |
| **Buckets** | `POST /buckets/:id/rotate-invite` | Rotate invite token. |
| **Buckets** | `GET /buckets/:id/members` | List members. |
| **Buckets** | `DELETE /buckets/:id/members/:user_id` | Revoke member. |
| **Join** | `POST /join` (body: invite_token or URL) | Resolve token → bucket_id; create bucket_member for current user; return bucket_id (and optionally owner/bucket name for alias suggestion). |
| **Items** | `POST /items` (body: bucket_id, title, optional due_date) | Add item; `source_user_id` = auth.uid(). |
| **Items** | `GET /items` (query: bucket_id, or “mine” for owner’s buckets; optional status, due_before, order) | List items (RLS: owner or member with read). |
| **Items** | `GET /items/contributed` | List items where source_user_id = auth.uid() (member view). |
| **Items** | `PATCH /items/:id` (body: status, optional decline_reason, completion_message, snooze_until) | Update item (owner only for status; RLS enforced). |

Listen (poll): `GET /items?bucket_id=...&updated_after=<timestamp>`. Realtime (post-MVP): Supabase Realtime on `items` table.

---

## 4. Database structure

### 4.1 Tables

**Users**  
- Use **Supabase Auth** `auth.users`; no separate `users` table required for MVP. If needed, a public `profiles` or `users` table keyed by `auth.uid()` for display name, etc.

**Buckets**

| Column | Type | Notes |
|--------|------|--------|
| `id` | UUID | PK, default gen_random_uuid(). |
| `owner_id` | UUID | FK to auth.users(id); NOT NULL. |
| `name` | TEXT | Bucket name (unique per owner). |
| `invite_token` | TEXT | Unique token for shareable link; nullable if only Model A. |
| `created_at` | TIMESTAMPTZ | Default now(). |

- **Unique:** (owner_id, name).
- **Index:** owner_id for “my buckets” queries.

**Bucket members**

| Column | Type | Notes |
|--------|------|--------|
| `id` | UUID | PK. |
| `bucket_id` | UUID | FK buckets(id) ON DELETE CASCADE. |
| `user_id` | UUID | FK auth.users(id). |
| `joined_at` | TIMESTAMPTZ | Default now(). |
| `invited_via` | TEXT | Optional: 'email' | 'token'. |

- **Unique:** (bucket_id, user_id).
- **Index:** user_id for “buckets I’m a member of”; bucket_id for “members of this bucket”.

**Items**

| Column | Type | Notes |
|--------|------|--------|
| `id` | UUID | PK. |
| `bucket_id` | UUID | FK buckets(id) ON DELETE CASCADE. |
| `source_user_id` | UUID | Who added it (auth.users). |
| `title` | TEXT | NOT NULL. |
| `status` | TEXT | pending | accepted | declined | snoozed | completed | archived; default 'pending'. |
| `due_date` | DATE | Optional. |
| `snooze_until` | TIMESTAMPTZ | Optional. |
| `decline_reason` | TEXT | Optional. |
| `completion_message` | TEXT | Optional. |
| `created_at` | TIMESTAMPTZ | Default now(). |
| `updated_at` | TIMESTAMPTZ | Default now(); update on change. |

- **Indexes:** bucket_id (list by bucket); (bucket_id, status, created_at) for list + filter; source_user_id for “contributed”; updated_at for listen poll.

**Invites (Model A — optional / post-MVP)**

| Column | Type | Notes |
|--------|------|--------|
| `id` | UUID | PK. |
| `bucket_id` | UUID | FK buckets. |
| `email` | TEXT | Invitee email. |
| `token` | TEXT | Unique token for accept link. |
| `status` | TEXT | pending | accepted | expired. |
| `created_at` | TIMESTAMPTZ | |

---

### 4.2 Row-level security (RLS)

- **Buckets:**  
  - SELECT: user is owner (owner_id = auth.uid()) OR user is in bucket_members for this bucket.  
  - INSERT: owner_id = auth.uid().  
  - UPDATE/DELETE: owner_id = auth.uid().

- **Bucket members:**  
  - SELECT: user is bucket owner OR user is the member (user_id = auth.uid()).  
  - INSERT: via join endpoint (invite token valid) or by owner (invite by email).  
  - DELETE: owner only (revoke).

- **Items:**  
  - SELECT: user is bucket owner OR user is in bucket_members (for that bucket) OR user is source_user_id (for “contributed” view).  
  - INSERT: user is bucket owner OR user is in bucket_members; set source_user_id = auth.uid().  
  - UPDATE: bucket owner only (status, decline_reason, completion_message, snooze_until).  
  - DELETE: bucket owner only (if we allow delete; else soft-delete via status).

Implement as Postgres policies on each table; Supabase enforces them on all access.

---

### 4.3 Invite token (Model B)

- **Storage:** `buckets.invite_token` — unique, indexed for fast lookup on join.
- **Join:** Request with invite_token → lookup bucket by invite_token → insert into bucket_members (user_id = auth.uid()) → return bucket_id (and bucket name, owner info for CLI alias suggestion).
- **Rotation:** Generate new invite_token for bucket; update row; old token no longer resolves (existing members unchanged).

---

## 5. Local storage (client-side)

The backend does **not** store per-user “remotes” or CLI config. Remotes (alias → bucket_id or token) and credentials are stored **only on the client** (see [cli-design.md § Local storage](cli-design.md#local-storage-local-db)). The backend only needs to:

- Resolve invite_token → bucket_id (and return it on join).
- Accept item create with bucket_id (member must be in bucket_members for that bucket).

No “remotes” or “aliases” table on the server for MVP.

---

## 6. Optional: Edge Functions

- **Join:** If join logic is more than “lookup bucket by token + insert bucket_member”, implement as Supabase Edge Function (e.g. validate token, create member, return bucket info).
- **Model A invite:** Sending email and accept-link handling typically need an Edge Function (or external service) to send email and to create bucket_member on accept.

---

## 7. References

| Doc | Content |
|-----|---------|
| [cli-design.md](cli-design.md) | CLI commands, remotes, local storage. |
| [invite-flow-and-scenarios.md](invite-flow-and-scenarios.md) | Invite models, member view, item lifecycle. |
| [research-and-stack.md](research-and-stack.md) | Stack choice, Supabase rationale. |
| [launch-requirements.md](launch-requirements.md) | MVP backend checklist. |

---

*Last updated: 2026-03-20 (Tray CLI + Go + Supabase naming)*
