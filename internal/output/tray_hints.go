package output

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

// WriteTrayHints prints suggested next commands after listing or creating trays (table output).
func WriteTrayHints(w io.Writer, trayNames []string) error {
	var b strings.Builder
	b.WriteString("Next steps:\n")
	if len(trayNames) == 1 {
		q := shellQuoteTrayName(trayNames[0])
		fmt.Fprintf(&b, "  tray add \"Task title\" %s\n", q)
		fmt.Fprintf(&b, "  tray invite %s\n", q)
	} else {
		b.WriteString("  tray add \"Task title\" <tray-name>\n")
		b.WriteString("  tray invite <tray-name>\n")
	}
	_, err := fmt.Fprint(w, b.String())
	return err
}

func shellQuoteTrayName(s string) string {
	if strings.ContainsAny(s, " \t\n\"'") {
		return strconv.Quote(s)
	}
	if s == "" {
		return `""`
	}
	return s
}
