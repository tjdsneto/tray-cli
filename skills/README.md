# Agent skills for the `tray` CLI

These files teach **Claude**, **Cursor**, and similar tools how to help **users** run the `tray` binary (auth, commands, hooks). They are **not** a substitute for [`docs/user/`](../docs/user/README.md)—they compress that material for LLM context.

## Contents

| Path | Role |
|------|------|
| [`tray-cli/SKILL.md`](tray-cli/SKILL.md) | **Canonical** skill body (version with the repo; suitable to zip or copy standalone). |

## Install without cloning (curl)

The skill file is plain markdown on GitHub. You can download it with **`curl`** and drop it into the tool’s skills directory—no `git clone` required.

**Stable raw URL (tracks `main`):**

`https://raw.githubusercontent.com/tjdsneto/tray-cli/main/skills/tray-cli/SKILL.md`

**Pin to a release** (matches a tagged release of the repo, e.g. `v1.2.3`):

`https://raw.githubusercontent.com/tjdsneto/tray-cli/v1.2.3/skills/tray-cli/SKILL.md`

Replace `v1.2.3` with the tag you want. If that path 404s, that release predates the skill file—use `main` or a newer tag.

### Cursor (personal skill)

Cursor loads project skills from **`.cursor/skills/`** inside a repo, and personal skills from **`~/.cursor/skills/`** (see Cursor’s skill docs). Example:

```bash
mkdir -p ~/.cursor/skills/tray-cli
curl -fsSL "https://raw.githubusercontent.com/tjdsneto/tray-cli/main/skills/tray-cli/SKILL.md" \
  -o ~/.cursor/skills/tray-cli/SKILL.md
```

### Claude Code (plugin marketplace)

The repo ships **`.claude-plugin/marketplace.json`** so you can add the marketplace from GitHub and install the **tray-cli** plugin (includes the skill):

```bash
claude plugin marketplace add tjdsneto/tray-cli
claude plugin install tray-cli@tjdsneto
```

When the repo updates on GitHub, refresh your local marketplace copy:

```bash
claude plugin marketplace update
```

For a **local clone** during development, you can use `claude plugin marketplace add .` from the repository root instead of the `tjdsneto/tray-cli` shorthand.

### Claude Code (personal skill, curl)

```bash
mkdir -p ~/.claude/skills/tray-cli
curl -fsSL "https://raw.githubusercontent.com/tjdsneto/tray-cli/main/skills/tray-cli/SKILL.md" \
  -o ~/.claude/skills/tray-cli/SKILL.md
```

**Use both tools?** You can keep one file and symlink the other to avoid drift, for example install under `~/.cursor/skills/tray-cli/SKILL.md` and run `ln -sf ~/.cursor/skills/tray-cli/SKILL.md ~/.claude/skills/tray-cli/SKILL.md` (adjust if you prefer Claude as the “source”).

## Keeping the skill updated

| Approach | What to do |
|----------|------------|
| **Claude Code plugin marketplace** | Run **`claude plugin marketplace update`** (after **`tray-cli`** changes are on GitHub). |
| **curl to `~/.claude/skills/…`** | Re-run the same **`curl`** (or your pinned-URL **`curl`**) so the file is overwritten. |
| **Track `main` with curl** | Re-run the install **`curl`** whenever you want the latest doc changes (same URL as above). |
| **Track releases with curl** | Set the URL to a **git tag** and bump the tag in the URL when you upgrade the CLI and want the skill in lockstep. |
| **Notifications** | Watch **Releases** on the GitHub repo, or subscribe to the repo’s RSS/activity, then update the marketplace or re-**`curl`** after a release if the release notes mention docs/skills. |

If a future **release tarball** ships `skills/tray-cli/SKILL.md`, you can unpack that file over your personal skill path instead of `curl`—see [`docs/maintainers/distribution.md`](../docs/maintainers/distribution.md).

## Cursor (project clone)

Cursor loads **project** skills from [`.cursor/skills/`](../.cursor/skills/). This repository symlinks **`tray-cli`** there to the canonical file so clones pick it up automatically.

To install **only** the skill in another project, copy `tray-cli/` (or symlink `SKILL.md`) into that project’s `.cursor/skills/tray-cli/`.

## Claude Code (project clone)

Project skills can live under **`.claude/skills/`**. This repo symlinks the same canonical `SKILL.md` so Claude Code sees it when opened at the repository root.

To use the skill elsewhere, copy `skills/tray-cli/SKILL.md` into `~/.claude/skills/tray-cli/SKILL.md` (personal) or another project’s `.claude/skills/tray-cli/SKILL.md`.

## Releases

Tarballs and install scripts are described in [`docs/maintainers/distribution.md`](../docs/maintainers/distribution.md). Optionally attach or document a small **`tray-cli-agent-skill.zip`** containing `tray-cli/SKILL.md` for users who do not clone the repo.
