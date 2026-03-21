package config

import "strings"

// EmbeddedDevOAuthHints is set to "1" at link time only for maintainer/debug binaries
// (see scripts/tray-env.sh: TRAY_EMBED_DEV_OAUTH_HINTS=1). Release builds leave it empty.
var EmbeddedDevOAuthHints = ""

// DevOAuthHintsEnabled reports whether this binary was built with maintainer-only OAuth diagnostics.
func DevOAuthHintsEnabled() bool {
	v := strings.TrimSpace(EmbeddedDevOAuthHints)
	return v == "1" || strings.EqualFold(v, "true") || strings.EqualFold(v, "yes")
}
