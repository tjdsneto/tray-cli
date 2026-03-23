package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/tjdsneto/tray-cli/internal/domain"
)

// WriteTrayMembers renders tray membership rows.
func WriteTrayMembers(w io.Writer, trayName string, members []domain.TrayMember, f Format) error {
	switch f {
	case FormatJSON:
		type out struct {
			Tray    string              `json:"tray"`
			Members []domain.TrayMember `json:"members"`
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(out{Tray: trayName, Members: members})
	case FormatMarkdown:
		if len(members) == 0 {
			_, err := fmt.Fprintf(w, "_No members listed for tray %s._\n", mdEscapeJoin(trayName))
			return err
		}
		_, err := fmt.Fprintf(w, "| user_id | joined_at | invited_via |\n|---|---|---|\n")
		if err != nil {
			return err
		}
		for _, m := range members {
			inv := ""
			if m.InvitedVia != nil {
				inv = strings.ReplaceAll(strings.ReplaceAll(*m.InvitedVia, "|", "\\|"), "\n", " ")
			}
			_, err := fmt.Fprintf(w, "| `%s` | %s | %s |\n",
				strings.ReplaceAll(m.UserID, "|", "\\|"),
				mdEscapeJoin(formatTrayLocalTime(m.JoinedAt)),
				inv)
			if err != nil {
				return err
			}
		}
		return nil
	default:
		if len(members) == 0 {
			_, err := fmt.Fprintf(w, "No members listed for tray %q.\n", trayName)
			return err
		}
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		_, err := fmt.Fprintln(tw, "USER_ID\tJOINED\tINVITED_VIA")
		if err != nil {
			return err
		}
		for _, m := range members {
			inv := ""
			if m.InvitedVia != nil {
				inv = *m.InvitedVia
			}
			_, err := fmt.Fprintf(tw, "%s\t%s\t%s\n", m.UserID, formatTrayLocalTime(m.JoinedAt), inv)
			if err != nil {
				return err
			}
		}
		return tw.Flush()
	}
}
