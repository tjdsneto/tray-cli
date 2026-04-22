package output

import (
	"io"
	"os"
	"regexp"
	"strings"
	"unicode/utf8"
)

var ansiSeq = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// StripANSI removes ANSI escape sequences (for width calculations).
func StripANSI(s string) string {
	return ansiSeq.ReplaceAllString(s, "")
}

// ColorEnabled is true for colorized human output: TTY stdout and NO_COLOR unset.
func ColorEnabled(w io.Writer) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	st, err := f.Stat()
	if err != nil {
		return false
	}
	return st.Mode()&os.ModeCharDevice != 0
}

const (
	ansiReset  = "\x1b[0m"
	ansiBold   = "\x1b[1m"
	ansiDim    = "\x1b[2m"
	ansiRed    = "\x1b[31m"
	ansiGreen  = "\x1b[32m"
	ansiYellow = "\x1b[33m"
	ansiBlue   = "\x1b[34m"
	ansiCyan   = "\x1b[36m"
)

// StatusSectionTitleANSI colors a tray list section heading (e.g. "Accepted", "Pending") to match status styling.
func StatusSectionTitleANSI(statusLower, title string, color bool) string {
	title = strings.TrimSpace(title)
	if title == "" {
		title = "—"
	}
	if !color {
		return title
	}
	var prefix string
	switch strings.ToLower(strings.TrimSpace(statusLower)) {
	case "pending":
		prefix = ansiYellow
	case "accepted":
		prefix = ansiGreen
	case "declined":
		prefix = ansiRed
	case "snoozed":
		prefix = ansiCyan
	case "completed":
		prefix = ansiGreen + ansiBold
	case "archived":
		prefix = ansiDim
	default:
		prefix = ansiBold
	}
	return prefix + title + ansiReset
}

// TrayGroupTitleANSI renders a tray name under a status section (highlighted when color is on).
func TrayGroupTitleANSI(trayName string, color bool) string {
	name := strings.TrimSpace(trayName)
	if name == "" {
		name = "—"
	}
	if !color {
		return name
	}
	return ansiBold + ansiBlue + name + ansiReset
}

// FormatStatusANSI returns a left-padded visual width string with optional ANSI color for known statuses.
func FormatStatusANSI(status string, color bool, width int) string {
	s := strings.TrimSpace(status)
	if s == "" {
		s = "—"
	}
	if !color {
		return padRightPlain(s, width)
	}
	prefix := ""
	switch strings.ToLower(s) {
	case "pending":
		prefix = ansiYellow
	case "accepted":
		prefix = ansiGreen
	case "declined":
		prefix = ansiRed
	case "snoozed":
		prefix = ansiCyan
	case "completed":
		prefix = ansiGreen + ansiBold
	case "archived":
		prefix = ansiDim
	default:
		return padRightPlain(s, width)
	}
	out := prefix + s + ansiReset
	n := utf8.RuneCountInString(s)
	if n < width {
		out += strings.Repeat(" ", width-n)
	}
	return out
}

func padRightPlain(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}
