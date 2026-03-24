package triageui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/tjdsneto/tray-cli/internal/domain"
	"github.com/tjdsneto/tray-cli/internal/output"
)

// triageItem is a list.DefaultItem for pending queue rows.
type triageItem struct {
	it          domain.Item
	trayNames   map[string]string
	displayByID map[string]string
	sess        domain.Session
}

var _ list.DefaultItem = triageItem{}

func (t triageItem) Title() string {
	return t.it.Title
}

func (t triageItem) Description() string {
	tn := t.trayNames[t.it.TrayID]
	if tn == "" {
		tn = t.it.TrayID
	}
	by := output.FormatSourceUser(t.it.SourceUserID, strings.TrimSpace(t.sess.UserID), t.displayByID)
	when := output.HumanizeTimeAgo(t.it.CreatedAt, time.Now())
	return fmt.Sprintf("%s · %s · %s", truncate(tn, 24), by, when)
}

func (t triageItem) FilterValue() string {
	return t.it.Title
}

func truncate(s string, max int) string {
	if max <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	if max == 1 {
		return "…"
	}
	return string(r[:max-1]) + "…"
}
