package output

import (
	"strings"
)

// FormatSourceUser renders who added an item: "you" for the current user, else a short id hint.
func FormatSourceUser(sourceUserID, currentUserID string) string {
	s := strings.TrimSpace(sourceUserID)
	if s == "" {
		return "—"
	}
	if strings.TrimSpace(currentUserID) != "" && strings.EqualFold(s, strings.TrimSpace(currentUserID)) {
		return "you"
	}
	// Short, stable hint without exposing full UUID in narrow columns.
	compact := strings.ReplaceAll(s, "-", "")
	if len(compact) > 8 {
		compact = compact[:8]
	}
	return compact
}
