# Naming ideas

Notes and candidates for product, features, and concepts.  
Inspired by: personal shared “radar” queue, non-urgent “put it on my radar,” CLI-first, agent-friendly.

**Decided (2026-03-20):** Official tool name **Tray CLI**; binary **`tray`**; implementation **Go**; backend **Supabase**. See [PRODUCT.md](PRODUCT.md) and [cli-design.md](cli-design.md). Main naming conflict: [Tray.io / Tray.ai](#tray--conflict-check).

**Earlier note:** Product metaphor **Tray** (inbox-tray); repo label **attention-queue**.

---

## CLI (binary name)

| Candidate | Notes |
|-----------|--------|
| **tray** | Matches product **Tray**; natural commands (`tray add`, `tray list work`). Unscoped npm `tray` is taken — ship via Homebrew / single binary; scoped npm if needed. |
| **aq** | Retired candidate (attention-queue); docs and CLI use **`tray`** / **Tray CLI** instead. |
| **ar** | Attention Radar — if you pivot naming to Radar. |

**Recommendation:** **`tray`** — aligns with the chosen product name.

---

## Product / app name

- **Tray** (working) — inbox tray; “add that to my tray”; things land for “when you get to it.” See [Tray conflict check](#tray--conflict-check).
- **Attention Queue** (repo / historical) — clear description of the thing.
- **Radar** — “put it on my radar”; single word, could conflict with other products.
- **Horizon** — things on your horizon; less literal.
- **Backlog (from others)** — accurate but “backlog” feels heavy.
- **Inbox (for later)** — inbox others can send to; “inbox” is overloaded.
- **Drop** / **Dropbox** — “drop it here for me”; Dropbox is taken.
- **Beacon** — “send me a beacon”; signal / ping vibe.
- **Scope** — “in scope for me”; one word.
- **Lens** — “through my lens” / what’s on my lens; a bit abstract.

### “Radar” — conflict check

If considering **Radar**, here’s what’s already using it:

| Use | Name | Space | Conflict? |
|-----|------|--------|-----------|
| Product | **RadaRist** (radar.ist) | Task/project management, “radar view” of projects | **Closest** — same metaphor (radar = overview), same broad space (tasks/productivity). Different product (their radar = bird’s-eye of projects; yours = “put it on my radar” queue). |
| Company / API | **Radar.io** | Geolocation, geofencing, maps | Different space (location). No direct conflict unless you go into geo. |
| npm package | **radar** (Zendesk) | Realtime/engine.io | **Taken** — so `radar` as an npm package name is gone. CLI binary can still be `radar` if you ship via Go/Rust/brew. |
| CLI | **Vault Radar** (HashiCorp) | Security scanning | Different (security). “Radar” as sub-product name. |
| App | **GetRadar** (getradar.co) | Menubar metrics | Different (dashboard/metrics). |

**Summary:** “Radar” as a **product/brand name** is still usable: RadaRist is the only close overlap (task/productivity + radar metaphor), but they’re “radar view of your projects” and you’re “put it on my radar” / shared queue — different angle. **CLI:** avoid claiming the unscoped npm name `radar`; ship binary **`tray`** (Tray CLI) via Homebrew/standalone (no npm), or use `ar` if you pivot naming to Radar. If you want to own a clear namespace: **OnRadar**, **MyRadar**, **Radar Queue**, or **Radar (attention queue)** in copy.

### No neutral term: product name = the term

**You don’t need “bucket.”** The product name can be the only concept:

- **“Add that to my tray”** / **“Add that to my radar”** — one phrase, one metaphor. The “thing” is your tray or your radar.
- **Segments** are just named instances: “my work tray,” “my life tray” or “my work radar,” “my life radar.” So the resource is **trays** or **radars**; each has a name (work, life, …). No second word.

That gives a **simple, non-confusing API and CLI**:

| Product | Resource (API) | Natural language | CLI example |
|---------|----------------|------------------|-------------|
| **Tray** | `trays` (each has name: work, life) | “Add to my tray” / “Add to my work tray” | `tray add "Review deck" work` |
| **Radar** | `radars` (each has name: work, life) | “Add to my radar” / “Add to my work radar” | `ar add "Review deck" work` |

**CLI shape (no “bucket”):** The segment is just the tray/radar name as an argument.

- `tray add "Review deck" work` → add to my work tray  
- `tray list work` → list items in my work tray  
- `tray invite work` → get invite link for my work tray  
- `tray members work` → who can add to my work tray  
- `tray rotate-invite work` → rotate invite for work tray  

Same for Radar with `ar` (or `radar`): `ar add "..." work`, `ar list work`, `ar invite work`, etc.

**API shape:** One resource type, keyed by name.

- `GET /trays` or `GET /radars` → my trays/radars  
- `POST /trays/work/items` or `POST /radars/work/items` → add item to work  
- `GET /trays/work/items` → list items in work  
- `POST /trays/work/invite` (or `GET /trays/work/invite-link`) → invite  

Backend can still use a table like `buckets` or `queues` internally; the **public** API and CLI only expose “tray” or “radar.” One term, one metaphor, no confusion.

---

### “Radar” + “bucket” (legacy note)

If you ever kept “bucket,” it could feel like two metaphors (radar + bucket). Dropping it and using “radar” as the only concept (named radars: work, life) avoids that — see “No neutral term” above.

### “Tray” — inbox tray (one metaphor)

**Office inbox tray** = the physical in-tray where papers landed for you to process. One clear metaphor: “put it in my tray” = non-urgent, when you get to it.

- **Product name: Tray** — “Add that to my tray”; “my work tray,” “my life tray.” No second term; the segment is just a named tray. CLI: `tray add "..." work`, `tray list work`, `tray invite work` (no “bucket”).
- **Domain:** See “Tray — conflict check” below; tray.com and tray.io are taken.

### “Tray” — conflict check

| Use | Name | Space | Conflict? |
|-----|------|--------|-----------|
| **Company / product** | **Tray.io** (rebranded **Tray.ai**) | iPaaS, workflow automation, AI agents, 700+ connectors; enterprise integration platform | **Major** — Same name “Tray,” well-funded ($149M+), Gartner Visionary. Different product (integration/automation, not inbox-tray productivity), but “Tray” is strongly associated with them. tray.io and tray.ai point to them. |
| **Company / product** | **Tray.com** | Enterprise POS (point of sale) for restaurants; IHOP, multi-unit operators | Different space (restaurant POS). Name “Tray” but not productivity/inbox. tray.com is taken. |
| **npm package** | **tray** (unscoped) | Node lib for system tray on OS X (brandonhorst/node-tray) | **Taken** — So `tray` as unscoped npm package is gone. CLI binary can still be `tray` if you ship via Go/Rust/brew. Use scoped npm if you ever ship a JS CLI (e.g. `@yourorg/tray-cli`). |
| **Other** | **Trayy** (Windows) | Minimize apps to system tray | Different spelling; no direct conflict. |
| **Other** | **ClipTray** (cliptray.app) | Clipboard manager, system tray | Compound name; “Tray” alone not claimed. |
| **Dev** | Electron `Tray` API | System tray in Electron apps | Generic API name; not a product. |

**Summary:** **Tray.io / Tray.ai** is the main conflict — same single-word name “Tray,” big presence in tech (integration/automation). Your product (inbox-tray / “add to my tray” for attention queue) is a different use case, but in search and word-of-mouth “Tray” will often mean them. **Options:** (1) Use **Tray** anyway and differentiate by tagline/positioning (“the inbox tray for your attention”) and avoid tray.io / tray.ai domains; (2) Use a compound: **InTray**, **MyTray**, **TrayInbox**, **OnTray** so the brand is distinct; (3) Prefer **Radar** or another name if you want a clearer namespace. For CLI: avoid unscoped npm `tray`; binary `tray` via Homebrew/standalone is fine.

**Strategy:** Decide product name first, then check availability. Often the name + TLD combo matters more than the TLD alone.

| TLD | Vibe | Notes |
|-----|------|--------|
| **.com** | Default, trusted | Ideal if available; for “Radar” / “Attention Queue” likely taken or pricey. Try compound names (e.g. **getradar.com**, **onradar.com**, **attentionqueue.com**). |
| **.io** | Tech, startup | Common for dev tools (e.g. radar.io is taken by the geo company). **youradar.io**, **myradar.io**, **radarqueue.io** — check availability. |
| **.co** | Short, startup | getradar.co exists (menubar app). **radar.co** may be taken; **onradar.co**, **myradar.co** possible. |
| **.app** | Web app | Requires HTTPS; good if the product has a web UI. e.g. **radar.app**, **attentionqueue.app**. |
| **.dev** | Developer | Google-run; fits CLI/dev-first products. e.g. **tray.dev** if you align domain with Tray CLI. |
| **.ist** | Niche | RadaRist uses radar.ist — so **radar.ist** likely taken. |
| **.list** / **.queue** | Thematic | Not in common use; check ICANN/registrars if they exist and are open for registration. |
| **.me** | Personal / “for me” | Montenegro ccTLD; open registration. **radar.me** = “my radar” / “radar for me” — strong semantic fit. Often better availability than .com. |
| **.so** | “Social” / software | Somalia ccTLD; used by dev tools. **radar.so** — check availability. |
| **.to** | Short / hacks | Tonga ccTLD; short, used for redirects and hacks (e.g. listen.to). **radar.to** possible. |
| **.today** | Timely | “What’s on your radar today.” **radar.today** — memorable, sometimes available. |
| **.link** | Connection | **radar.link** — “link to my radar.” |
| **.run** | Execution | Fits CLI/dev; **radar.run** — “run your radar.” |

### Strong “good name + often available” combos

- **radar.me** — “My radar” / “radar for me”; .me is real, open, and the phrase fits the product. **Top candidate** to check.
- **onradar.me**, **getradar.me**, **myradar.me** — compounds with .me if radar.me is taken.
- **radar.so**, **radar.to**, **radar.today**, **radar.link**, **radar.run** — alternative TLDs that can read well and are often less crowded than .com/.io.

**Practical tips:**
- **Compound name + .com** (e.g. **getradar.com**, **onmyradar.com**) often more available than single word.
- **.io** and **.dev** read well for a CLI-first product.
- Once you have a shortlist (e.g. Radar, OnRadar, AQ), run quick checks on [namecheap](https://www.namecheap.com), [Google Domains](https://domains.google), or [Cloudflare Registrar](https://www.cloudflare.com/products/registrar/) before locking the brand.

**Not a TLD:** **.go** is not a real TLD (no `.go` in the DNS root; two-letter TLDs are country codes and there is no country “go”). So **radar.go** is not a registerable domain. Use .com, .io, .dev, etc. instead.

---

## Related words (for copy, concepts, or future names)

**Queue / flow:** queue, stream, pipeline, funnel, channel, feed  
**Attention / visibility:** attention, radar, horizon, scope, lens, visibility, signal  
**Non-urgent / async:** when you get to it, async, non-urgent, low-pressure, digest, triage  
**Sharing / contributors:** drop, add, contribute, invite, shared, “for me”  
**Containers:** bucket, list, inbox, backlog  
**Actions:** review, prioritize, triage, process, clear, drain  

Use these for CLI help text, docs, or taglines (e.g. “Your queue for what’s on your radar”).
