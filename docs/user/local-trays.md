# Local trays (directory, branch, global)

Tray can store items **on this machine** without sign-in. Use these for project context, branch WIP, and personal globals. **Remote trays** (`tray create`, `tray add --remote`) stay on the server for handoffs and collaboration.

## Tray kinds

| Kind | Identity | Created when |
|------|----------|--------------|
| **Global** | Name, e.g. `inbox` | First `tray add "…" inbox` |
| **Directory** | Absolute path of a folder | First add with `--here`, `--project`, or default on default branch |
| **Branch** | Git repo root + branch name | First add with `--branch`, or default on a non-`main`/`master` branch |

Trays are registered in `~/.config/tray/local/index.json` (or `$TRAY_CONFIG_DIR/local/`). Item bodies live alongside as jsonl files under `local/items/`.

## Default scope: `tray list` and `tray ls`

Without flags, both commands show **local trays in context**:

1. **Branch tray** for the current git branch (if it exists in the index)
2. **Directory trays** for the current working directory and each parent path (indexed only)
3. **Global** local trays (always listed if they exist)

Nothing is created until you add an item (create-on-add).

| Flag | Effect |
|------|--------|
| **`--all`** | Every local tray ever created on this machine |
| **`--remote`** | Remote trays you own (`tray list --remote`, `tray ls --remote`; requires sign-in) |

## Adding items

```bash
tray add "Fix the flaky test"              # branch tray if not on main/master, else cwd dir tray
tray add "Note" inbox                      # local global tray "inbox"
tray add "WIP" --here                      # directory tray for cwd
tray add "WIP" --branch                    # branch tray for current branch
tray add "WIP" --project                   # directory tray for git repo root
tray add "Hand off" work --remote          # remote tray (requires sign-in)
tray add "Note" inbox --no-create          # fail if local tray does not exist yet
```

Local **`tray complete`** and **`tray remove`** work on local item ids (from `tray list`) without sign-in.

## Prune empty context trays

```bash
tray prune              # drop empty directory/branch trays from the index
tray prune --dry-run    # preview only
```

Global trays are kept even when empty.

## Remote vs local

- **Local add** — always stored in local files; no approval flow.
- **Remote add** (`--remote`) — server tray; pending when contributing to someone else's tray (unchanged).
- Promoting a directory tray to remote (invite others) is not in v1 yet; remote remains the collaboration layer.

## Agents

When entering a repo, run **`tray list`** (and **`tray ls`**) to see context trays and open items. File branch or project notes with **`tray add "…" --branch`** or **`--here`** as needed.
