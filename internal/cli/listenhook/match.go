package listenhook

import (
	"strings"

	"github.com/tjdsneto/tray-cli/internal/domain"
)

// OwnedTraySet is tray id -> true when the current user owns the tray.
func OwnedTraySet(trays []domain.Tray, userID string) map[string]bool {
	uid := strings.TrimSpace(userID)
	out := make(map[string]bool)
	for _, t := range trays {
		tid := strings.TrimSpace(t.ID)
		if tid == "" {
			continue
		}
		if strings.TrimSpace(t.OwnerID) == uid {
			out[tid] = true
		}
	}
	return out
}

// MatchPending returns whether a hook rule should run for this new pending item.
// trayFilter is optional resolved tray id; empty means any tray in owned set.
func MatchPending(r Rule, it domain.Item, owned map[string]bool, sessionUserID string, trayFilter string) bool {
	if strings.TrimSpace(r.Event) != EventItemPending {
		return false
	}
	if pendingScope(&r) != ScopeInboxOwned {
		return false
	}
	tid := strings.TrimSpace(it.TrayID)
	if tid == "" || !owned[tid] {
		return false
	}
	if !strings.EqualFold(strings.TrimSpace(it.Status), "pending") {
		return false
	}
	tf := strings.TrimSpace(trayFilter)
	if tf != "" && tid != tf {
		return false
	}
	if r.FromOthersDefault() {
		if strings.TrimSpace(it.SourceUserID) == strings.TrimSpace(sessionUserID) {
			return false
		}
	}
	return true
}

// MatchOutboxFilter returns whether the rule's outbox scope and optional tray filter match the item.
// The caller must already match r.Event to the hook event being fired.
func MatchOutboxFilter(r Rule, it domain.Item, trayFilter string) bool {
	if outboxScope(&r) != ScopeOutbox {
		return false
	}
	tf := strings.TrimSpace(trayFilter)
	if tf == "" {
		return true
	}
	return strings.TrimSpace(it.TrayID) == tf
}
