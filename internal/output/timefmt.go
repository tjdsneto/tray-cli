package output

import "time"

// formatTrayLocalTime renders a timestamp in the process local timezone for human tables.
func formatTrayLocalTime(t time.Time) string {
	return t.In(time.Local).Format("Jan 2, 2006 3:04 PM MST")
}
