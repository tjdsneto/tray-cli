package cli

import "strings"

// UserFacingError rewrites a few common cobra messages so stderr is easier to read.
func UserFacingError(err error) string {
	if err == nil {
		return ""
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
