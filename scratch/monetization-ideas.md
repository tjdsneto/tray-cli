# Monetization ideas (scratch)

Working notes on **freemium**, **paid levers**, and how that interacts with **public vs private repo**. Not product commitments—iterate as you learn from users.

**Product context:** CLI-first shared “tray” / attention queue (Tray working name); hosted backend (e.g. Supabase); others add items via invite link; owner triages.

---

## 1. Freemium: what to charge for (when you’re stuck)

The free tier should feel **complete for a solo person or a tiny circle** so word-of-mouth works. Paid should unlock **scale, durability, teams, or polish**—not “the app barely works without paying.”

### A. Natural paid dimensions (fit this product)

| Lever | Free (example) | Paid (example) | Why it’s fair |
|-------|----------------|----------------|---------------|
| **Trays (buckets)** | 3 trays | Unlimited or higher cap | Power users with many life areas pay. |
| **Members per tray** | 5 members / tray | 50+ or unlimited | Heavy collaboration = value. |
| **Items / month** | Soft cap (e.g. 500 adds/mo) or fair use | Higher limits | Prevents abuse; scales with usage. |
| **Retention / history** | Last 90 days visible; older archived or hidden | Longer or “forever” + export | Storage + value for audit trail. |
| **API / CLI rate limits** | Generous for humans | Higher for automation / agents | Teams piping agents in pay. |
| **`listen` / realtime** (post-MVP) | Poll only or low frequency | Webhooks, faster poll, push | “Notify me when something lands” is pro. |
| **Webhooks / integrations** | None or 1 | Slack, email digest, Zapier | Clear B2B upsell. |
| **Team / org billing** | Personal only | Seats, central admin, SSO later | Standard SaaS ladder. |
| **Branding on invite links** | Generic subdomain | Custom domain for join URLs | Vanity + trust for companies. |
| **Support** | Community / docs | Email SLA, priority | Obvious at scale. |
| **Compliance / export** | Basic export | Audit logs, bulk export, DPA | Enterprise track. |

Many of these **don’t exist in MVP yet**—that’s fine. You can launch **100% free** (or “free while in beta”) and add metering once you see real usage patterns.

### B. Things that are *weak* as first paid features

- **“More colors” / cosmetic only** — unless you have a strong consumer brand.
- **Core triage (accept/decline/complete)** — gating basics kills trust for a “simple tray” product.
- **OAuth login** — charging for “sign in with GitHub” feels hostile; keep auth free.

### C. Models beyond classic freemium

| Model | Fit |
|-------|-----|
| **Open core** | CLI (and maybe client libs) open source; **hosted sync** is the paid product. Aligns with “trust the CLI, pay for convenience.” |
| **Sponsorware / donations** | GitHub Sponsors, optional paid tier for “supporter” badge—fine as **side** income, rarely enough alone. |
| **Lifetime / early-bird** | One-time for first N users—cash + goodwill; watch support expectations. |
| **Usage-based** | Per active member or per 1k items—fair for B2B; more billing complexity. |
| **Seat-based (teams)** | $/user/month for org features—simple story for companies. |

**Practical MVP stance:** Ship **free hosted MVP** (or invite-only). Add **Stripe + one clear limit** (e.g. trays or members) when you have **10+ serious users** and a sense of who hits limits.

---

## 2. Analogous products (how others implement similar monetization)

These aren’t identical to Tray, but they show **patterns** that map to the levers in §1—useful when you design tiers or explain them to others.

### A. Open / local client + paid **hosted** service (**open core**)

| Product | What’s free | What’s paid | Parallel to Tray |
|---------|-------------|-------------|------------------|
| **Obsidian** | Full local app; community plugins | **Sync** (encrypted cloud), **Publish** (hosted sites) | CLI/local usage free; **hosted queue + sync** is the subscription—same story as “pay for cloud, not for the binary.” |
| **Standard Notes** | Local notes | **Sync** subscription | Same: convenience + multi-device via their servers. |
| **LogSeq** (historically) | Open-source, local-first | Optional **hosted sync** (when offered) | Hosted graph/sync as paid add-on on top of open client. |

**Takeaway:** Strong precedent for “developer-trusty client is free / open; **reliable hosted backend** is the product you charge for.”

### B. **Free tier + limits** on usage, history, or scale

| Product | Free tier hook | Paid unlock | Parallel to Tray |
|---------|------------------|-------------|------------------|
| **Slack** (workspace) | Full product for small use | **Message history** beyond ~90 days, more apps, huddles at scale | Maps to **retention / history** on items and “serious team” upsell. |
| **Notion** | Personal / small team | **Guests**, **blocks** / file upload limits, **SAML** on business | Maps to **members per tray**, **items/volume**, **team/org** features. |
| **Discord** | Full chat | **Nitro** (cosmetics + uploads); **Server boosts** | Weaker parallel for *features*, but shows **optional paid** on top of a free social graph. |
| **Supabase** | Generous free project | **DB size**, **MAU**, **egress**, add-ons | Same *infra* you might use—your **marginal cost** scales like theirs; you pass limits through as paid tiers. |
| **PlanetScale** | Free DB branches | **Row/read limits**, branches, support | Classic **usage caps** → paid when hobby becomes production. |
| **ngrok** | Tunnels with session / connection limits | **Reserved domains**, more agents, SSO, team dashboard | Maps to **CLI + connectivity**: free for individuals; **custom domains / team / reliability** paid—like **branded invite links** + org features. |

**Takeaway:** “Free until you’re clearly extracting value” via **history**, **collaborator count**, **API/egress**, or **pro infrastructure** (domains, SSO).

### C. **Seat-based** and **team** SaaS (collaboration products)

| Product | Model | Parallel to Tray |
|---------|--------|------------------|
| **Linear** | Free for small teams; paid **per seat** with admin, SAML, etc. | **Members across trays**, audit, SSO when you go upmarket. |
| **Figma** | Free starter; **paid seats** for full team libraries / permissions | Same seat story for “many people adding to my trays.” |
| **Postman** | Free; **team workspaces** and governance paid | Collaboration + **rate/usage** on shared team usage—like shared buckets and API-heavy agent use. |
| **1Password** | Individual vs **Families** vs **Business** (per seat) | Simple ladder: solo → household → org. |

**Takeaway:** Once Tray is **multi-person by design**, **per-seat** or **per-workspace** billing is a familiar buyer story.

### D. **Developer tools** with a **Pro** lane (not only hosted)

| Product | Free | Pro | Parallel to Tray |
|---------|------|-----|------------------|
| **Raycast** | Core launcher + extensions | **Pro**: AI, cloud sync of settings/snippets | Optional **automation / agent-adjacent** features behind Pro—like **`listen`**, **webhooks**, higher **API limits**. |
| **GitHub** | Public repos, basic Actions | **Copilot**, **Advanced Security**, larger **Actions** for private orgs | Mixed: some paid is **features**, some is **scale**—both apply as you grow. |

**Takeaway:** You can combine **hosted scale** (§A–B) with **power-user / automation** upsells (§D) without confusing the core free story.

### E. Quick mapping cheat sheet

| Your lever (§1) | See also |
|-----------------|----------|
| Hosted sync / trust in cloud | Obsidian Sync, Standard Notes |
| History / retention | Slack free vs paid history |
| Members, guests, seats | Notion, Linear, Figma, Postman teams |
| Usage / DB / API limits | Supabase, PlanetScale, ngrok tiers |
| Custom domain / “pro” tunnel | ngrok reserved domains |
| Open client + paid server | Obsidian, LogSeq-style models |

*Pricing and SKUs change—check each vendor’s current page when you benchmark numbers, not just the pattern.*

---

## 3. Does monetization push you toward a **private** repo?

**Not automatically.** They’re separate decisions:

| Repo choice | Monetization |
|-------------|--------------|
| **Public** | Very common for **open core**: code visible, money from **hosted service**, support, or **managed** tiers. Competitors can fork; your moat is **product, UX, speed, and billing**. |
| **Private** | Common for **closed-source SaaS**. You still distribute the **CLI binary** publicly (separate public releases repo or CDN—see [repo-visibility-and-distribution.md](repo-visibility-and-distribution.md)). |

**Freemium does not require a private repo.** Many freemium devtools are public on GitHub; payment is tied to **account limits on your servers**, not to hiding the client.

**When private helps:**

- You don’t want competitors copying **exact** backend integration or pricing experiments in real time.
- You want **cleaner** separation between “our IP” and “community forks.”

**When public helps:**

- Trust, security review, and **CLI adoption** (especially for agents and developers).
- Simpler story: “the client is open; you pay for cloud.”

**Hybrid:** Public **CLI** repo + private **backend / ops** repo—or public monorepo with proprietary pieces only on the server (typical for open core).

---

## 4. Decision prompts (fill in later)

- [ ] MVP: **100% free**, **beta badge**, or **waitlist + invite**?
- [ ] First paid dimension when you add billing: **trays**, **members**, or **retention**?
- [ ] Target customer first: **individual knowledge worker** vs **small team** vs **org**?
- [ ] Open core vs closed source: **\_\_\_\_\_\_**

---

## 5. References

- [PRODUCT.md](../PRODUCT.md) — positioning  
- [launch-requirements.md](../docs/launch-requirements.md) — MVP scope  
- [repo-visibility-and-distribution.md](../docs/repo-visibility-and-distribution.md) — public vs private + distribution  

*Scratch note — 2026-03-18 (§2: analogous products).*
