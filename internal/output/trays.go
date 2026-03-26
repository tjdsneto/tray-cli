package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/tjdsneto/tray-cli/internal/domain"
)

// trayAccess returns "owner" vs "member" for the signed-in user, or empty if viewerUserID is unset.
func trayAccess(t domain.Tray, viewerUserID string) string {
	v := strings.TrimSpace(viewerUserID)
	if v == "" {
		return ""
	}
	if strings.TrimSpace(t.OwnerID) == v {
		return "owner"
	}
	return "member"
}

// WriteTrays renders trays for list-style commands. When showHints is true and format is
// table, prints suggested commands after the table (create / ls only).
// When viewerUserID is non-empty, adds an ACCESS column (owner = you own the tray; member = you joined someone else's tray).
func WriteTrays(w io.Writer, trays []domain.Tray, f Format, showHints bool, viewerUserID string) error {
	showAccess := strings.TrimSpace(viewerUserID) != ""
	switch f {
	case FormatJSON:
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(trays)
	case FormatMarkdown:
		if len(trays) == 0 {
			_, err := fmt.Fprint(w, "_No trays yet._\n\n_Create a tray:_ `tray create <name>`\n")
			return err
		}
		if showAccess {
			_, err := fmt.Fprintf(w, "| %s | %s | %s | %s |\n", "Name", "Access", "Items", "Created")
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(w, "| %s | %s | %s | %s |\n", "---", "---", "---", "---")
			if err != nil {
				return err
			}
			for _, t := range trays {
				name := strings.ReplaceAll(t.Name, "|", "\\|")
				_, err := fmt.Fprintf(w, "| %s | %s | %d | %s |\n", name, trayAccess(t, viewerUserID), t.ItemCount, formatTrayLocalTime(t.CreatedAt))
				if err != nil {
					return err
				}
			}
			return nil
		}
		_, err := fmt.Fprintf(w, "| %s | %s | %s |\n", "Name", "Items", "Created")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(w, "| %s | %s | %s |\n", "---", "---", "---")
		if err != nil {
			return err
		}
		for _, t := range trays {
			name := strings.ReplaceAll(t.Name, "|", "\\|")
			_, err := fmt.Fprintf(w, "| %s | %d | %s |\n", name, t.ItemCount, formatTrayLocalTime(t.CreatedAt))
			if err != nil {
				return err
			}
		}
		return nil
	default:
		if len(trays) == 0 {
			if showHints {
				_, err := fmt.Fprint(w, "No trays yet.\n\nCreate your first tray:\n  tray create <name>\n")
				return err
			}
			_, err := fmt.Fprintln(w, "No trays.")
			return err
		}
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		if showAccess {
			_, err := fmt.Fprintln(tw, "NAME\tACCESS\tITEMS\tCREATED")
			if err != nil {
				return err
			}
			for _, t := range trays {
				_, err := fmt.Fprintf(tw, "%s\t%s\t%d\t%s\n", t.Name, trayAccess(t, viewerUserID), t.ItemCount, formatTrayLocalTime(t.CreatedAt))
				if err != nil {
					return err
				}
			}
		} else {
			_, err := fmt.Fprintln(tw, "NAME\tITEMS\tCREATED")
			if err != nil {
				return err
			}
			for _, t := range trays {
				_, err := fmt.Fprintf(tw, "%s\t%d\t%s\n", t.Name, t.ItemCount, formatTrayLocalTime(t.CreatedAt))
				if err != nil {
					return err
				}
			}
		}
		if err := tw.Flush(); err != nil {
			return err
		}
		if !showHints {
			return nil
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
		names := make([]string, len(trays))
		for i := range trays {
			names[i] = trays[i].Name
		}
		return WriteTrayHints(w, names)
	}
}
