# Agent skills for the `tray` CLI

These files teach **Claude**, **Cursor**, and similar tools how to help **users** run the `tray` binary (auth, commands, hooks). They are **not** a substitute for [`docs/user/`](../docs/user/README.md)—they compress that material for LLM context.

## Contents

| Path | Role |
|------|------|
| [`tray-cli/SKILL.md`](tray-cli/SKILL.md) | **Canonical** skill body (version with the repo; suitable to zip or copy standalone). |

## Cursor (project clone)

Cursor loads **project** skills from [`.cursor/skills/`](../.cursor/skills/). This repository symlinks **`tray-cli`** there to the canonical file so clones pick it up automatically.

To install **only** the skill in another project, copy `tray-cli/` (or symlink `SKILL.md`) into that project’s `.cursor/skills/tray-cli/`.

## Claude Code (project clone)

Project skills can live under **`.claude/skills/`**. This repo symlinks the same canonical `SKILL.md` so Claude Code sees it when opened at the repository root.

To use the skill elsewhere, copy `skills/tray-cli/SKILL.md` into `~/.claude/skills/tray-cli/SKILL.md` (personal) or another project’s `.claude/skills/tray-cli/SKILL.md`.

## Releases

Tarballs and install scripts are described in [`docs/maintainers/distribution.md`](../docs/maintainers/distribution.md). Optionally attach or document a small **`tray-cli-agent-skill.zip`** containing `tray-cli/SKILL.md` for users who do not clone the repo.
