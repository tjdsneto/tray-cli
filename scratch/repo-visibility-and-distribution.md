# Repository visibility: public vs private (GitHub Releases & Pages)

How **public vs private** affects **GitHub Releases**, **GitHub Pages**, and **one-liner install scripts**. Use this when deciding whether the main repo should be public.

---

## 1. Do Releases and Pages require a public repo?

### GitHub Releases

- **Public repo:** Release assets have **public download URLs**. Anyone can `curl` a binary or use an install script **without** logging into GitHub. Homebrew taps and CI can fetch artifacts anonymously.
- **Private repo:** Release assets are only available to users with **read access** to the repository. Anonymous downloads **do not** work. A public install script that points at `github.com/.../releases/download/...` will **fail** for strangers unless you embed a token (don’t do that for end users).

**Implication:** Open distribution (random users install the CLI) almost always means either a **public** repo for releases **or** binaries hosted somewhere else with a public URL (e.g. R2, S3, Cloudflare).

### GitHub Pages

Per [GitHub’s plans](https://docs.github.com/en/get-started/learning-about-github/githubs-plans):

- **GitHub Free (personal):** GitHub Pages is available for **public** repositories only. You **cannot** publish Pages from a **private** repo on the free plan.
- **GitHub Pro / Team / Enterprise:** You can publish Pages from a **private** repository. The **published site** at `*.github.io` is still **public** by default (anyone can open the URL); only the **source** repo stays private. (Enterprise can restrict site access in some setups.)

**Implication:** On a **free** account, a **minimal auth success page** on GitHub Pages either needs a **public** repo (often a tiny `docs` or `pages` repo) **or** you host that page elsewhere (Netlify, Vercel, Cloudflare Pages, Supabase-hosted redirect, etc.).

---

## 2. Patterns if you want a private main repo

| Pattern | What you do |
|---------|-------------|
| **Public “releases only” repo** | Private monorepo for product; separate **small public** repo whose only job is Releases + `install.sh` (CI copies binaries from private build). |
| **Public static site repo** | Private code; separate **public** repo with only the auth callback / success HTML for Pages (or use Netlify/Vercel free tier). |
| **Public object storage** | Build in private repo; upload artifacts to R2/S3/Cloudflare with **public read**; install script and Homebrew point there. |
| **Paid GitHub** | Keep repo private; use Pages from private repo + public release URLs… **Note:** Releases from a **private** repo are still **not** anonymously downloadable—so you still need a public release host or a public releases repo for open installs. |

---

## 3. Pros and cons of a **public** main repo

### Pros

| Pro | Why it matters |
|-----|----------------|
| **Simple distribution** | Public Releases = anonymous install scripts, Homebrew, no extra infra. |
| **Free GitHub Pages** | Auth success / docs site from the same repo on a free account. |
| **Trust and adoption** | Users and security reviewers can inspect the CLI source; common expectation for developer tools. |
| **Contributions** | Issues, PRs, and forks without granting private access. |
| **No duplicate repos** | One place for code, issues, releases, and optional Pages. |

### Cons

| Con | Why it matters |
|-----|----------------|
| **Source is visible** | Implementation, API hints, and history are open; competitors can read them. |
| **Everything public** | Issues, discussions, and (by default) Actions logs may expose more than you want unless you’re careful. |
| **Less “stealth”** | Harder to keep the product invisible before launch (you can still avoid marketing it). |
| **Perceived obligation** | Some teams feel pressure to maintain docs, semver, and community once public. |

### Pros of **private** main repo (brief)

- Roadmap, messy early history, and internal notes stay off the public internet.
- You can still ship a public CLI using a **public releases repo** or **public CDN** (see §2).

### Cons of **private** main repo for a **public** CLI

- You must **explicitly** solve hosting for **anonymous** binary downloads (second public repo, or object storage).
- On **GitHub Free**, you **cannot** use Pages from that private repo; use a separate public pages repo or another host.

---

## 4. Recommendation (MVP)

- **Easiest path:** **Public** repo + GitHub Releases + optional GitHub Pages (auth success page) on a free account.
- **If you want private code:** Plan for a **small public** artifact surface (releases repo and/or static site) or **public CDN** for binaries; keep product docs that reference those URLs.

---

*Related: [research-and-stack.md](research-and-stack.md) (CLI distribution), [launch-requirements.md](launch-requirements.md) (§3.1 Distribution, §3.2 Auth).*

*Last updated: 2026-03-18*
