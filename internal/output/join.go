package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// WriteJoin prints the result of joining a tray (human, json, or markdown).
func WriteJoin(w io.Writer, trayID, name string, f Format) error {
	switch f {
	case FormatJSON:
		type out struct {
			TrayID string `json:"tray_id"`
			Name   string `json:"name,omitempty"`
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(out{TrayID: trayID, Name: strings.TrimSpace(name)})
	case FormatMarkdown:
		n := name
		if n == "" {
			n = "—"
		}
		_, err := fmt.Fprintf(w, "| | |\n|--|--|\n| tray_id | `%s` |\n| name | %s |\n", mdEscapeJoin(trayID), mdEscapeJoin(n))
		return err
	default:
		if n := strings.TrimSpace(name); n != "" {
			_, err := fmt.Fprintf(w, "Joined tray %q.\n", n)
			if err != nil {
				return err
			}
		} else {
			_, err := fmt.Fprintf(w, "Joined the tray (run `tray ls` to see it — id %s).\n", trayID)
			if err != nil {
				return err
			}
		}
		_, err := fmt.Fprint(w, "\nNext steps:\n  tray ls\n  tray add \"Task title\" <tray-name>\n")
		return err
	}
}

func mdEscapeJoin(s string) string {
	if s == "" || s == "—" {
		return s
	}
	return strings.ReplaceAll(strings.ReplaceAll(s, "|", "\\|"), "\n", " ")
}
