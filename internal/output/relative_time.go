package output

import (
	"fmt"
	"time"
)

// HumanizeTimeAgo renders a friendly relative time from t to now (e.g. "20 minutes ago").
func HumanizeTimeAgo(t, now time.Time) string {
	if t.After(now) {
		return t.In(time.Local).Format("Jan 2, 2006 3:04 PM")
	}
	d := now.Sub(t)
	if d < time.Minute {
		return "just now"
	}
	if d < time.Hour {
		m := int(d / time.Minute)
		if m == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", m)
	}
	if d < 24*time.Hour {
		h := int(d / time.Hour)
		if h == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", h)
	}
	if d < 48*time.Hour {
		return "yesterday"
	}
	if d < 7*24*time.Hour {
		days := int(d / (24 * time.Hour))
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
	return t.In(time.Local).Format("Jan 2, 2006 3:04 PM")
}
