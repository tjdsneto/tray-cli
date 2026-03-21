package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/tjdsneto/tray-cli/internal/domain"
)

// WriteTrays renders trays for list-style commands. When showHints is true and format is
// table, prints suggested commands after the table (create / ls only).
func WriteTrays(w io.Writer, trays []domain.Tray, f Format, showHints bool) error {
	switch f {
	case FormatJSON:
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(trays)
	case FormatMarkdown:
		if len(trays) == 0 {
			_, err := fmt.Fprintln(w, "_No trays._")
			return err
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
			_, err := fmt.Fprintln(w, "No trays.")
			return err
		}
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
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
