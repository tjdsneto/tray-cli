# Privacy policy — Tray and ChatGPT Actions (gateway)

**Effective date:** 2026-05-05  
**Repository:** [github.com/tjdsneto/tray-cli](https://github.com/tjdsneto/tray-cli)

This policy describes how **Tray** (shared inbox / attention queues) and the optional **ChatGPT Custom GPT** integration backed by the **Tray HTTP gateway** handle personal data. Tray is open source; the gateway runs as a **Supabase Edge Function** and uses **Supabase Auth** and database **row-level security (RLS)** so each user only accesses their own permitted data.

## Who this applies to

- Users of the **Tray CLI** and related services.
- Users who **connect** a Custom GPT to Tray via **OAuth** (ChatGPT Actions).

If you only use Tray without ChatGPT, the ChatGPT-specific sections do not apply.

## Data we process

### Account and identity

- When you sign in (CLI `tray login` or gateway OAuth), we use **Supabase Auth** and your chosen identity provider (for example **Google**). We process identifiers and profile information that provider supplies for authentication.
- The gateway may store **OAuth connection state** and **short-lived gateway access tokens** (and related refresh material) **server-side** so API calls can run in your name. Sensitive material is handled per the gateway design (for example encrypted at rest, hashed identifiers where appropriate). Details for operators are in [`docs/maintainers/chatgpt-gateway.md`](maintainers/chatgpt-gateway.md).

### Tray content

- **Trays, items, memberships, and triage** you create or that others file for you are stored in the Tray backend. Access is enforced by **RLS** and your role (owner, member, contributor) — not by the GPT reading your screen.

### ChatGPT and the model

- **OpenAI / ChatGPT** may process **conversation text** you send to the Custom GPT according to OpenAI’s policies.
- When you use **Actions**, ChatGPT sends **API requests** to the Tray gateway with an **authorization token** tied to **your** connected account. The assistant does not receive your Supabase refresh token; gateway tokens are short-lived and scoped (for example `tray.read`, `tray.write`).

### Logs and operations

- The service may keep **technical logs** (for example request metadata, errors, latency) for security and reliability. Those logs are not sold and are used to operate the service.

## What we do not do

- We do **not** use your tray content to train third-party models (Tray does not operate a public model training product).
- We do **not** expose one user’s trays or items to another user except where you already share access (for example trays you joined or items you contributed by design).

## Your choices

- **Disconnect** the Custom GPT in ChatGPT when you no longer want it to call the gateway on your behalf.
- **Revoke** application access through your identity provider or Supabase/session flows as documented for your setup.
- For **CLI** users, you can sign out or remove local credentials per [`docs/user/`](user/README.md) and CLI help.

## Children

Tray is not directed at children under 13 (or the minimum age in your jurisdiction). Do not use the service if you are not old enough to agree to these terms in your region.

## Changes

We may update this page when the product or gateway changes. The **effective date** at the top will be revised; substantive changes are reflected in the repository history of this file.

## Contact

For privacy questions about this open-source project, open an issue or discussion on **[github.com/tjdsneto/tray-cli](https://github.com/tjdsneto/tray-cli)** or contact the maintainers there. This document is informational; it is not individualized legal advice.

---

**ChatGPT “Privacy policy” URL (after this file is on `main`):**  
`https://github.com/tjdsneto/tray-cli/blob/main/docs/privacy.md`
