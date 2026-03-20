# Cost analysis (scratch)

Rough **monthly** and **one-time** costs for **MVP phase** vs **“real product” phase** (**Tray CLI**, repo attention-queue, **Supabase** stack). Numbers are **order-of-magnitude**—verify against current vendor pricing before budgeting.

**Assumptions:** Small team (1–2 people), OAuth-only MVP (no transactional email), EU/US acceptable latency.

---

## 1. Cost categories

| Category | MVP phase | Real product phase |
|----------|-----------|---------------------|
| Backend (DB + Auth + API) | Usually **free tier** or low paid | Scales with MAU, DB size, bandwidth |
| Domains & TLS | **$0–15/yr** if skipped | **~$10–50/yr** + optional email |
| Email (transactional) | **$0** if OAuth-only | **$0–50+/mo** (Resend, Postmark, etc.) |
| Binary / static hosting | **$0** (GitHub Releases + Pages or R2 free tier) | **$0–30/mo** if you add CDN or egress-heavy downloads |
| CI / builds | **$0** (GitHub Actions free for public/small private) | **$0–100+/mo** at scale |
| Billing (Stripe) | **$0** until you charge | **~2.9% + $0.30** per successful card charge; no big fixed fee |
| Legal / compliance (rough) | **$0–500** one-time templates | **$1k–10k+/yr** as you add DPAs, terms, review |
| Support & ops | Your time | Tools (e.g. $20–100/mo) optional |

---

## 2. Backend: Supabase (illustrative)

Supabase (and similar) bill on **database size**, **monthly active auth users**, **egress**, **Edge Function** invocations, etc. Tiers change—check [supabase.com/pricing](https://supabase.com/pricing).

| Phase | Typical picture | Ballpark |
|-------|-----------------|----------|
| **MVP** | Few dozen users, tiny DB, low egress | **$0/mo** (free tier) if you stay within limits |
| **MVP stress** | Beta with hundreds of users, moderate API traffic | **$25–50/mo** (first paid tier) |
| **Real product** | Thousands of MAU, growth, backups, maybe read replicas later | **$25–300+/mo** depending on size; enterprise custom |

**Watchouts:** Accidental **egress** (large `list` responses, log dumps), **Realtime** if you add it, **Edge Function** volume on hot paths.

---

## 3. GitHub

| Item | MVP | Real product |
|------|-----|--------------|
| **Repo (private)** | Free for individuals; org features extra | Same + possible **Copilot**, **Actions** minutes if private + heavy CI |
| **Actions** | Usually enough for release builds | Buy packs if you exceed free minutes |
| **Pages** | Free if public source repo (see repo-visibility doc) | Pro if Pages from private repo (~$4/user/mo) |

---

## 4. Domain & DNS

| Item | Cost |
|------|------|
| **Domain** | ~**$10–20/year** (.com-ish) |
| **DNS** | Often included at registrar or free (Cloudflare) |

MVP can stay on **Supabase project URL** only ($0).

---

## 5. Email

| Phase | Cost |
|-------|------|
| **MVP (OAuth-only, link invites)** | **$0** |
| **Magic link / password reset / invite-by-email** | Supabase included quota or **Resend/Postmark** ~**$10–40/mo** at low volume |

---

## 6. Artifacts: CLI binaries & install script

| Approach | MVP cost | Notes |
|----------|----------|--------|
| **GitHub Releases (public)** | $0 | Simplest |
| **R2 / S3 + public read** | ~$0–5/mo at small scale | Egress pricing if huge download volume |
| **Cloudflare R2** | Often negligible at MVP | Good if repo is private |

---

## 7. Payment processing (when you monetize)

| Provider | Typical |
|----------|---------|
| **Stripe** | No monthly minimum; **2.9% + $0.30** (US cards); Stripe Tax / Billing extras optional |
| **Apple/Google** | If you ever do mobile IAP—different fee structure |

Budget **~3–4%** of revenue as payment + failed payment overhead until you optimize.

---

## 8. Rough monthly totals (summary)

| Phase | Lean scenario | “Comfortable” scenario |
|-------|---------------|-------------------------|
| **MVP** | **$0–10/mo** (free Supabase + GitHub + no domain) | **$25–40/mo** (paid Supabase tier + domain amortized) |
| **Real product (early)** | **$50–150/mo** (DB + email + domain + Stripe) | **$200–500+/mo** (traffic, more email, monitoring, legal amortized) |

**One-time MVP:** **$0–500** (domain, optional template legal docs).

---

## 9. What to track before you optimize costs

- **MAU** (*monthly active users* — e.g. distinct users who authenticate or use the API in a month) vs Supabase tier thresholds  
- **DB size** (items + indexes)  
- **Egress** per user (CLI polling `listen` will matter later)  
- **Support time** (often the real “cost” before infra hurts)

---

## 10. References

- [monetization-ideas.md](./monetization-ideas.md) — when to add paid tiers  
- [launch-requirements.md](../docs/launch-requirements.md) — MVP technical scope  
- [repo-visibility-and-distribution.md](../docs/repo-visibility-and-distribution.md) — hosting binaries with private repo  

*Scratch note — 2026-03-18 — replace ballparks with a spreadsheet when you pick vendors.*
