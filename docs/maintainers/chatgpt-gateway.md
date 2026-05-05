# ChatGPT gateway plan (gateway-issued tokens)

This plan adds a public API surface for ChatGPT Actions while keeping tray data access user-specific.

**Operator how-to (GPT builder UI):** [`chatgpt-custom-gpt.md`](chatgpt-custom-gpt.md).

## Goal

Support a **single shared GPT** where each person connects their own account, and requests run as that user.

High-level shape:

1. ChatGPT Action uses **OAuth** to connect a user.
2. Gateway issues a short-lived **gateway access token** (not a Supabase refresh token).
3. API calls include `Authorization: Bearer <gateway_token>`.
4. Gateway resolves user identity and calls tray data paths with user-scoped access.

## Non-goals (v1)

- Do not expose raw PostgREST tables/filters as the public contract.
- Do not store long-lived secrets in the GPT instruction body.
- Do not depend on one static API key shared by all users of the GPT.

## Architecture

Components:

- **Public gateway API** (Edge Function or small service): stable, OpenAPI-described endpoints.
- **OAuth endpoints** (authorize + token): used by ChatGPT Action auth flow.
- **Token/session store**: stores gateway refresh/session material server-side (encrypted at rest).
- **Supabase adapter**: uses existing tray data model and RLS behavior through user context.

Trust boundaries:

- ChatGPT only sees gateway tokens needed for API calls.
- Supabase refresh/session details remain server-side.
- Gateway enforces scopes/rate limits/audit per connected user.

## Auth model

### 1) ChatGPT <-> gateway (OAuth)

Use Action OAuth for per-user identity:

- User clicks **Connect** in ChatGPT.
- Gateway OAuth returns access token (+ refresh token if needed).
- ChatGPT stores per-user token and sends bearer token on each action call.

### 2) Gateway token format

Use short-lived signed JWT (or opaque token + lookup), with:

- `sub` (gateway user id)
- `scope` (e.g. `tray.read`, `tray.write`, `tray.triage`)
- `exp` (short TTL, e.g. 5-15 minutes)
- `iss`, `aud`

Refresh logic stays at the gateway token endpoint; API routes validate access token only.

### 3) Gateway <-> Supabase

Gateway calls Supabase with user-scoped context:

- Preferred: pass/derive user access context so RLS remains authoritative.
- If service-role operations are needed (e.g. auth/session bookkeeping), isolate them to internal tables and never bypass tray authorization checks.

## API surface (v1)

Start with small endpoints for assistant workflows:

- `GET /v1/me` - connected user identity + scopes.
- `GET /v1/trays` - owned/joined trays (human-oriented summary).
- `GET /v1/trays/{tray}/items` - list items with minimal filters (`status`, `limit`, `cursor`).
- `POST /v1/trays/{tray}/items` - add item.
- `PATCH /v1/items/{id}` - limited triage actions.

Keep responses stable and task-oriented; hide PostgREST internals from callers.

## Data model additions (gateway side)

Add tables for integration auth state (names may vary):

- `gateway_oauth_clients` (if multi-client support)
- `gateway_user_connections` (`user_id`, provider identity, status, created/updated)
- `gateway_refresh_tokens` (hashed token id, encrypted payload, expiry, revoked metadata)
- `gateway_audit_logs` (user id, route, result, latency, request id)

Security rules:

- Store only hashes for bearer token identifiers where possible.
- Encrypt sensitive token payloads at rest.
- Support key rotation for signing/encryption.

## Rollout plan

### Completed spike (not kept in repo)

- Validated that ChatGPT can reach a Supabase-hosted HTTPS endpoint.
- Validated that Action-shaped auth headers and request flow work end-to-end.
- Removed spike code/config from this repository to keep focus on production OAuth implementation.

### Phase 1: OAuth bootstrap

- Implement `/oauth/authorize` and `/oauth/token`.
- Register ChatGPT callback URI allowlist (gateway enforces `GATEWAY_ALLOWED_REDIRECT_PREFIXES`).
- Redirect-only login: `/oauth/authorize` → Supabase Auth OAuth (PKCE, same pattern as `tray login`) → `/oauth/supabase-callback/{pending_id}` → ChatGPT `redirect_uri` with code.
- Return gateway-issued access token for connected user.
- Add `GET /v1/me` protected endpoint.

**Supabase Dashboard (required):**

- **Authentication → URL configuration → Redirect URLs:** allow  
  `https://<project-ref>.supabase.co/functions/v1/gateway/oauth/supabase-callback/**`  
  (or the exact callback URL pattern your project uses; wildcards are supported per [Redirect URLs](https://supabase.com/docs/guides/auth/redirect-urls)).
- **Authentication → Providers:** enable the same OAuth provider the gateway uses (default **Google**), matching your existing `tray login` setup.

**Gateway secrets / env (typical):**

- `GATEWAY_OAUTH_PROVIDER` — default `google`.
- `GATEWAY_ALLOWED_CLIENT_IDS` — comma-separated (default includes `chatgpt-gateway-dev`).
- `GATEWAY_ALLOWED_REDIRECT_PREFIXES` — comma-separated HTTPS prefixes (defaults include ChatGPT hosts).

Exit criteria:

- One shared GPT connects different users independently.
- `GET /v1/me` returns different identities per user.

### Phase 2: read paths

- Persist Supabase Auth access/refresh tokens in `gateway_user_supabase_auth` during the PKCE callback so PostgREST runs as the same user (RLS matches the tray CLI). Refresh on demand when the cached access token is near expiry.
- Implement `GET /v1/trays` (owned only, `tray ls`), `GET /v1/remotes` (joined, `tray remote ls`), `GET /v1/items` (all visible items / optional `tray_id`, `status`, `limit`), `GET /v1/items/contributed` (`tray contributed`), and `GET /v1/trays/{tray}/items`.
- Add pagination and deterministic ordering.
- Add request IDs and structured logs.

Exit criteria:

- GPT can reliably list trays/items for each user.
- No cross-user data leakage in tests.

### Phase 3: write paths

- Implemented: `POST /v1/trays/{trayId}/items`, `DELETE /v1/items/{itemId}`, and triage `POST` actions
  `complete`, `accept`, `decline`, `snooze`, `archive` under `/v1/items/{itemId}/…`. All require OAuth scope **`tray.write`** (in addition to `tray.read` for reads).
- Remaining: idempotency keys and per-route rate limits for mutating routes.

Exit criteria:

- Mutations are safe to retry and audited.
- Abuse controls are active.

### Phase 4: production hardening

- SLOs/alerts, dead-letter handling for transient failures.
- Token rotation and revocation UX.
- Security review (threat model + pen-test checklist).

## Operational checklist

- **Privacy policy URL (public GPTs / Actions):** host or link to [`docs/privacy.md`](../privacy.md) — canonical GitHub view: `https://github.com/tjdsneto/tray-cli/blob/main/docs/privacy.md` (must be on `main` after merge).
- HTTPS + stable domain for OpenAPI `servers`.
- OAuth callback URLs configured for ChatGPT domains.
- Versioned OpenAPI docs (`v1` paths; additive changes preferred).
- Error contract with stable machine-readable codes.
- Audit trail for connect/disconnect/token refresh and data mutations.

## Risks and mitigations

- **Token leakage in logs** -> redact auth headers and secrets everywhere.
- **Over-broad scopes** -> default to read-only; request write scopes explicitly.
- **Replay/retry duplicates** -> require idempotency keys on writes.
- **RLS bypass risk** -> keep user-context data paths separate from service-role internals.

## Testing strategy

- Unit tests for token mint/verify and scope checks.
- Integration tests for OAuth callback and token refresh.
- Authorization tests: user A cannot access user B trays/items.
- Contract tests for OpenAPI examples and error envelope stability.
