# Creating the Tray Custom GPT (ChatGPT Actions)

**Audience:** maintainers and integrators who publish a **Custom GPT** that talks to the **Tray gateway**. This is **not** end-user documentation; CLI users do not need this file.

**Related:** architecture and rollout — [`chatgpt-gateway.md`](chatgpt-gateway.md). **Privacy policy URL** for public GPTs — [`docs/privacy.md`](../privacy.md).

## Prerequisites

1. **Gateway deployed** on Supabase (Edge Function `gateway`) with OAuth and JWT secrets set in the dashboard (see [`.env.example`](../../.env.example) comments and [`chatgpt-gateway.md`](chatgpt-gateway.md) Phase 1).
2. **Supabase Auth** — same OAuth provider as production (often **Google**); **Redirect URLs** include the gateway callback pattern documented in `chatgpt-gateway.md`.
3. **`GATEWAY_ALLOWED_CLIENT_IDS`** includes the **Client ID** you will enter in ChatGPT (for example `chatgpt-gateway-dev` or your own id).
4. **`GATEWAY_ALLOWED_REDIRECT_PREFIXES`** includes ChatGPT origins (for example `https://chat.openai.com/`, `https://chatgpt.com/`).
5. **OpenAPI `servers`** in [`supabase/functions/gateway/openapi.yaml`](../../supabase/functions/gateway/openapi.yaml) use your real gateway base URL, e.g.  
   `https://<project-ref>.supabase.co/functions/v1/gateway`  
   Re-import the schema in ChatGPT after changing it.

## 1. Create the GPT

In **ChatGPT** → **Create** → configure **Name**, **Description**, and **Instructions** so the model knows it is helping with **Tray** (trays, items, triage). Keep instructions **behavioral** (how to list trays, resolve names to IDs, handle errors); do **not** paste secrets or raw JWTs.

Suggested **conversation starters**: list trays, list items, add an item, triage actions — whatever matches your audience.

## 2. Actions — import the schema

1. Open **Configure** → **Actions** → **Create new action** (or edit existing).
2. Import **Schema**: paste the contents of `openapi.yaml` or provide a **public URL** to the raw file if you host it (GitHub raw URL to `main` works if the repo is public).
3. Confirm **Authentication** is set to **OAuth** (next section), not API key.

After import, ChatGPT should show operations such as `listTrays`, `listAllItems`, `addTrayItem`, triage routes, etc.

## 3. Actions — OAuth (Client credentials)

Use the gateway’s OAuth endpoints (same host as `servers` in the OpenAPI file):

| Field | Value |
|--------|--------|
| **Authorization URL** | `https://<project-ref>.supabase.co/functions/v1/gateway/oauth/authorize` |
| **Token URL** | `https://<project-ref>.supabase.co/functions/v1/gateway/oauth/token` |
| **Client ID** | Must match an allowed id in `GATEWAY_ALLOWED_CLIENT_IDS` |
| **Client secret** | Leave empty unless your deployment uses a confidential client (default dev flow often has no secret). |
| **Scope** | `tray.read tray.write` — required if the GPT should **add**, **delete**, or **triage** items; use `tray.read` only for read-only. |

ChatGPT will append standard OAuth query parameters (`response_type`, `redirect_uri`, `state`, etc.). The gateway implements the authorize + token exchange described in the OpenAPI paths `/oauth/authorize` and `/oauth/token`.

**User experience:** the end user clicks **Connect** in the GPT, signs in with the configured provider, and ChatGPT stores tokens for subsequent Action calls.

## 4. Privacy policy (required for public GPTs)

OpenAI requires a **valid privacy policy URL** when the GPT (or its Actions) are **public**.

- Use the published policy:  
  `https://github.com/tjdsneto/tray-cli/blob/main/docs/privacy.md`  
  (after that file exists on `main` in the public repo.)

Paste it into the GPT’s **Privacy policy** (or equivalent) field in **Configure**.

## 5. Publish vs private

- **Private / link-only:** fewer policy checks; good for testing with a small group.
- **Public:** privacy URL must validate; OAuth redirect allowlists and stable `servers` URL matter more.

## 6. Troubleshooting

| Symptom | What to check |
|---------|----------------|
| Connect fails or blank redirect | Gateway logs; Supabase **Redirect URLs**; `GATEWAY_ALLOWED_REDIRECT_PREFIXES`. |
| Reads work, writes return 403 | User reconnect with scope **`tray.read tray.write`**; gateway JWT scope claims. |
| Wrong trays / counts | Gateway lists **owned** trays on `listTrays`; **joined** trays use `listRemotes` — align instructions with the API. |
| Stale schema | Re-upload or re-fetch `openapi.yaml` after server or path changes. |

## Why this lives under `docs/maintainers/`

Publishing a Custom GPT involves **your** Supabase project, **your** OAuth allowlists, and **OpenAI’s** product UI. That is operator and contributor work, not something we put in **`docs/user/`** for people who only install the CLI binary.
