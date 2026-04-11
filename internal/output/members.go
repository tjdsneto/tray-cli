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
// displayByID maps user_id → profile name or email (from LookupDisplay); nil uses short id hints via FormatSourceUser.
func WriteTrayMembers(w io.Writer, trayName string, members []domain.TrayMember, currentUserID string, displayByID map[string]string, f Format) error {
	switch f {
	case FormatJSON:
		type memberJSON struct {
			domain.TrayMember
			UserLabel string `json:"user_label"`
		}
		type out struct {
			Tray    string       `json:"tray"`
			Members []memberJSON `json:"members"`
		}
		rows := make([]memberJSON, 0, len(members))
		for _, m := range members {
			rows = append(rows, memberJSON{
				TrayMember: m,
				UserLabel:  FormatSourceUser(m.UserID, currentUserID, displayByID),
			})
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(out{Tray: trayName, Members: rows})
	case FormatMarkdown:
		if len(members) == 0 {
			_, err := fmt.Fprintf(w, "_No members listed for tray %s._\n", mdEscapeJoin(trayName))
			return err
		}
		_, err := fmt.Fprintf(w, "| name | user_id | joined_at | invited_via |\n|---|---|---|---|\n")
		if err != nil {
			return err
		}
		for _, m := range members {
			inv := ""
			if m.InvitedVia != nil {
				inv = strings.ReplaceAll(strings.ReplaceAll(*m.InvitedVia, "|", "\\|"), "\n", " ")
			}
			label := FormatSourceUser(m.UserID, currentUserID, displayByID)
			_, err := fmt.Fprintf(w, "| %s | `%s` | %s | %s |\n",
				mdEscapeJoin(label),
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
		_, err := fmt.Fprintln(tw, "NAME\tUSER_ID\tJOINED\tINVITED_VIA")
		if err != nil {
			return err
		}
		for _, m := range members {
			inv := ""
			if m.InvitedVia != nil {
				inv = *m.InvitedVia
			}
			name := FormatSourceUser(m.UserID, currentUserID, displayByID)
			_, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", name, m.UserID, formatTrayLocalTime(m.JoinedAt), inv)
			if err != nil {
				return err
			}
		}
		return tw.Flush()
	}
}
