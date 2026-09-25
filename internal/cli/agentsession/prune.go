package agentsession

import (
	"strings"
	"time"

	"github.com/tjdsneto/tray-cli/internal/domain"
)

const DefaultPruneIdle = 7 * 24 * time.Hour

func IsOpenStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "pending", "accepted", "snoozed":
		return true
	default:
		return false
	}
}

// Prunable reports whether an agent-session tray may be deleted: correct name,
// no open items, and last activity (tray UpdatedAt or CreatedAt, or latest item UpdatedAt) older than idle.
func Prunable(tray domain.Tray, items []domain.Item, now time.Time, idle time.Duration) bool {
	if !IsAgentSessionTrayName(tray.Name) {
		return false
	}
	last := tray.UpdatedAt
	if last.IsZero() {
		last = tray.CreatedAt
	}
	for _, it := range items {
		if IsOpenStatus(it.Status) {
			return false
		}
		if it.UpdatedAt.After(last) {
			last = it.UpdatedAt
		}
	}
	if idle <= 0 {
		idle = DefaultPruneIdle
	}
	return !last.IsZero() && now.Sub(last) >= idle
}
