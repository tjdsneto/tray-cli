# Tray-CLI

**Tray** is a **CLI** for shared **inbox trays**: add items to your tray or someone else’s, **triage** what lands on yours, and keep **asks and handoffs** from dissolving into chat scrollback.

**Chat and DMs** work in the moment—but “I’ll get to that” is easy to **forget**, and threads get **buried** under whatever came next. Tray gives each owner a **persistent queue**: items stay on **your tray** until **you** act on them; people you **invite** can still file requests without a separate project tool.

Built with **Go** and **Supabase**. **Try it:** [Install](#install) → **`tray login`** → **`tray create`** / **`tray invite`**. See **[`docs/user/hooks.md`](docs/user/hooks.md)** if you want **`tray listen`** and local notifications.

**More documentation:** **[`docs/user/`](docs/user/README.md)** (hooks index) · **[`docs/maintainers/`](docs/maintainers/README.md)** (contributors) · [`docs/README.md`](docs/README.md) (full index) · **[`skills/`](skills/README.md)** (Claude/Cursor: operating `tray`).

## Install

**From GitHub Releases:**

```bash
curl -fsSL https://raw.githubusercontent.com/tjdsneto/tray-cli/main/scripts/install.sh | bash
```

**Where it installs:** the script **does not run `sudo` unless you set `TRAY_INSTALL_USE_SUDO=1`**. It picks the first **writable** directory: reuse the path if `tray` is already on `PATH`, else **`/usr/local/bin`** or **`/opt/homebrew/bin`** (macOS) when your user can write there, else **`~/.local/bin`**. That last path is often not on `PATH` on macOS until you add it (the installer prints copy-paste steps). System-wide install without write access: `TRAY_INSTALL_USE_SUDO=1 TRAY_INSTALL_DIR=/usr/local/bin` (one password prompt). Override directory with `TRAY_INSTALL_DIR`.

**Upgrades:** run **`tray upgrade`** (install-script method). To pin: **`tray upgrade --version v0.1.0`**. You can still run the same `curl … | bash` line directly with default **`TRAY_VERSION=latest`**.

**With Go** (needs Go 1.25+):

```bash
go install github.com/tjdsneto/tray-cli/cmd/tray@latest
```

Ensure `$(go env GOPATH)/bin` is on your `PATH`.

Versioning, release tarballs, and publishing are documented for maintainers in **[`docs/maintainers/distribution.md`](docs/maintainers/distribution.md)**.

## Configuration

Config directory: override with **`TRAY_CONFIG_DIR`**; otherwise **Windows** uses `%APPDATA%\tray`, **macOS/Linux** use `$XDG_CONFIG_HOME/tray` if set, else `~/.config/tray` (see `internal/config/paths.go`).

**Supabase:** `TRAY_SUPABASE_URL` (e.g. `https://xxxx.supabase.co`), `TRAY_SUPABASE_ANON_KEY`. At runtime, **environment variables override** values embedded at build time. See [`.env.example`](.env.example).

## Login & session

Run **`tray login`** to open a **local web page** to sign in with **Google** (enable it in Supabase). Use **`tray login --provider <id>`** or **`TRAY_OAUTH_PROVIDER`** in `.env` to skip the page and use a specific provider your project has enabled. If you already have a **valid saved session**, the CLI skips the browser until **`tray login --force`**. After **OAuth**, the CLI **refreshes the access JWT** using the stored refresh token when it is expired or near expiry. **`tray login --token`** stores only an access token—use OAuth for automatic refresh.

**Status:** **`tray status`** checks credentials and validates the session with Supabase (`--format json` for scripts; exit code **0** if signed in, **1** if not).

During OAuth, the CLI starts a **short-lived local HTTP server** on `127.0.0.1` with a **random free port** (`:0`) so the browser can return the auth code—that is normal; it is not listening on the network.

**Google / GitHub OAuth apps** (outside Supabase): **Authorized redirect URI** = **`https://<project-ref>.supabase.co/auth/v1/callback`**. Tokens are written to **`credentials.json`** under the config directory.

**Login (manual token):** `tray login --token '<access_jwt>'` — validates via `GET /auth/v1/user` and writes credentials (no browser).

**Troubleshooting OAuth:** If you see `Unsupported provider: provider is not enabled`, open **Supabase Dashboard → Authentication → Providers**, turn the provider on, and paste **Client ID** / **Client secret**. The CLI cannot enable providers from the client.

## Trays

**`tray create <name>`** creates a tray; **`rename <tray> <new-name>`** (owner); **`delete-tray`** (owner); **`ls`** lists **only trays you own**; **`join <token-or-url> [local-alias]`** joins via invite; **`invite`** / **`rotate-invite`** manage invite tokens (owner); **`members`**, **`revoke`**, **`leave`** for membership.

**Joined trays** (someone else’s inbox you were invited to) show up in **`tray remote ls`**, with optional local aliases — see **[`docs/user/trays.md`](docs/user/trays.md)**.

## Items

**`tray add "title" <tray>`** adds a pending item (tray = name, id, or **`remote`** alias). **`list`** shows items on **your** trays; **`list <tray>`** requires an **owned** tray. **`contributed`** lists items you filed on others’ trays; **`remove <item-id>`** (owner deletes any item; contributor can delete own **pending** items).

## Triage (tray owner)

**`accept`**, **`decline`** (**`--reason`**), **`snooze`** (**`--until` RFC3339**), **`complete`** (**`--message`**), **`archive`**. Use item ids from **`tray list --format json`**.

## Listen hooks

**`tray listen`** can run **`hooks.json`** when items change (pending on your trays, outbox accepted/declined/completed). See **[`docs/user/hooks.md`](docs/user/hooks.md)** for events, environment variables, and recipes.

## Remote aliases

**`join … <alias>`** or **`tray remote add <alias> <invite-url-or-token>`** saves `remotes.json`. **`tray remote rename <current> <new>`** renames an alias. **`tray remote ls`** / **`tray remote remove <alias>`** manage that file.

## Output (list-style commands)

**Default is human-friendly:** tables, local dates, and contextual hints where we’ve added them.

| Flag | Purpose |
|------|---------|
| **`--format human`** (default) | Friendly tables and hints |
| **`--format json`**, **`--format machine`**, or **`--json`** | Stable JSON for scripts |
| **`--format markdown`** / **`md`** | Markdown tables |

**Deprecated but still works:** `-o` / `--output` (same values as `--format`). Prefer **`--format`**.

`--json` is shorthand for `--format json` and must not be combined with another explicit format.

For **trays**, the default **human** output shows **name**, **item count**, and **created** (local timezone; set **`TZ`** if needed). Tray **IDs** and **`item_count`** appear in **`--format json`**.

For **items** (`list`, `review`, …), human output includes **who added** the item (`you` vs a short id), **created** as a **relative time** when recent, and **status** colors on a TTY. Set **`NO_COLOR=1`** to disable ANSI colors. **`tray triage`** is an interactive pending queue (TTY); **`tray review`** is non-interactive.

## Debugging

**`TRAY_DEBUG=1`** prints full PostgREST response bodies when something fails. By default, errors are shortened for readability.

---

## Maintainers and contributors

Everything for **developing this repo**—tests, local builds, releases, migrations, architecture—is in **[`docs/maintainers/README.md`](docs/maintainers/README.md)**. Start there if you are cloning the source, running **`./run.sh`**, or cutting a release.
