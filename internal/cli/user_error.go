package cli

import (
	"errors"
	"strings"

	"github.com/tjdsneto/tray-cli/internal/cli/errs"
)

// UserFacingError rewrites a few common cobra messages so stderr is easier to read.
func UserFacingError(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, errs.MissingBackendConfig) {
		return "Tray isn’t configured to reach your server yet. Set TRAY_SUPABASE_URL and TRAY_SUPABASE_ANON_KEY, or use a build from your team that already includes them. With TRAY_DEBUG=1, the raw error is printed above."
	}
	s := err.Error()
	switch {
	case strings.HasPrefix(s, "unknown command "):
		return "That isn't a tray command — run `tray help` to see what's available.\n\n(" + s + ")"
	case strings.HasPrefix(s, "unknown shorthand flag"):
		return "That option isn't recognized — try `tray help` or `tray <command> --help`.\n\n(" + s + ")"
	case strings.HasPrefix(s, "unknown flag"):
		return "That option isn't recognized — try `tray help` or `tray <command> --help`.\n\n(" + s + ")"
	}
	return s
}
