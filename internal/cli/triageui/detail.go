package triageui

import (
	"fmt"
	"strings"
	"time"

	"github.com/tjdsneto/tray-cli/internal/domain"
	"github.com/tjdsneto/tray-cli/internal/output"
)

func formatRFC3339Ptr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

// detailLines returns lines for the selected item (timestamps + metadata for triage context).
func detailLines(it domain.Item, trayNames map[string]string, displayByID map[string]string, sess domain.Session) []string {
	tn := trayNames[it.TrayID]
	if tn == "" {
		tn = it.TrayID
	}
	by := output.FormatSourceUser(it.SourceUserID, strings.TrimSpace(sess.UserID), displayByID)
	now := time.Now()
	lines := []string{
		fmt.Sprintf("Tray: %s", tn),
		fmt.Sprintf("Order: %d", it.SortOrder),
		fmt.Sprintf("By:   %s", by),
		fmt.Sprintf("Created: %s (%s)", output.HumanizeTimeAgo(it.CreatedAt, now), it.CreatedAt.UTC().Format(time.RFC3339)),
	}
	if it.DueDate != nil && strings.TrimSpace(*it.DueDate) != "" {
		lines = append(lines, fmt.Sprintf("Due:    %s", strings.TrimSpace(*it.DueDate)))
	}
	if s := formatStatusLine("Accepted:", it.AcceptedAt); s != "" {
		lines = append(lines, s)
	}
	if s := formatStatusLine("Declined:", it.DeclinedAt); s != "" {
		lines = append(lines, s)
	}
	if s := formatStatusLine("Completed:", it.CompletedAt); s != "" {
		lines = append(lines, s)
	}
	if s := formatStatusLine("Archived:", it.ArchivedAt); s != "" {
		lines = append(lines, s)
	}
	if s := formatStatusLine("Snoozed:", it.SnoozedAt); s != "" {
		lines = append(lines, s)
	}
	return lines
}

func formatStatusLine(label string, t *time.Time) string {
	if t == nil {
		return ""
	}
	return fmt.Sprintf("%s %s", label, formatRFC3339Ptr(t))
}
