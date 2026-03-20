package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/tjdsneto/tray-cli/internal/domain"
)

// WriteTrays renders trays for list-style commands.
func WriteTrays(w io.Writer, trays []domain.Tray, f Format) error {
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
		_, err := fmt.Fprintf(w, "| %s | %s | %s |\n", "Name", "ID", "Created")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(w, "| %s | %s | %s |\n", "---", "---", "---")
		if err != nil {
			return err
		}
		for _, t := range trays {
			name := strings.ReplaceAll(t.Name, "|", "\\|")
			_, err := fmt.Fprintf(w, "| %s | `%s` | %s |\n", name, t.ID, t.CreatedAt.UTC().Format(time.RFC3339))
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
		_, err := fmt.Fprintln(tw, "NAME\tID\tCREATED")
		if err != nil {
			return err
		}
		for _, t := range trays {
			_, err := fmt.Fprintf(tw, "%s\t%s\t%s\n", t.Name, t.ID, t.CreatedAt.UTC().Format(time.RFC3339))
			if err != nil {
				return err
			}
		}
		return tw.Flush()
	}
}
