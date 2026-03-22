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

// WriteItems renders items; trayNames maps tray_id → display name (optional).
func WriteItems(w io.Writer, items []domain.Item, trayNames map[string]string, f Format) error {
	if trayNames == nil {
		trayNames = map[string]string{}
	}
	switch f {
	case FormatJSON:
		type row struct {
			ID           string  `json:"id"`
			TrayID       string  `json:"tray_id"`
			TrayName     string  `json:"tray_name,omitempty"`
			Title        string  `json:"title"`
			Status       string  `json:"status"`
			DueDate      *string `json:"due_date,omitempty"`
			CreatedAt    string  `json:"created_at"`
			SourceUserID string  `json:"source_user_id"`
		}
		out := make([]row, 0, len(items))
		for _, it := range items {
			out = append(out, row{
				ID:           it.ID,
				TrayID:       it.TrayID,
				TrayName:     trayNames[it.TrayID],
				Title:        it.Title,
				Status:       it.Status,
				DueDate:      it.DueDate,
				CreatedAt:    it.CreatedAt.UTC().Format(time.RFC3339),
				SourceUserID: it.SourceUserID,
			})
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(out)
	case FormatMarkdown:
		if len(items) == 0 {
			_, err := fmt.Fprintln(w, "_No items._")
			return err
		}
		_, err := fmt.Fprintf(w, "| %s | %s | %s | %s |\n", "Status", "Title", "Tray", "Created")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(w, "| %s | %s | %s | %s |\n", "---", "---", "---", "---")
		if err != nil {
			return err
		}
		for _, it := range items {
			tn := trayNames[it.TrayID]
			if tn == "" {
				tn = it.TrayID
			}
			tn = strings.ReplaceAll(tn, "|", "\\|")
			ttl := strings.ReplaceAll(it.Title, "|", "\\|")
			_, err := fmt.Fprintf(w, "| %s | %s | %s | %s |\n",
				strings.ReplaceAll(it.Status, "|", "\\|"),
				ttl,
				tn,
				formatTrayLocalTime(it.CreatedAt))
			if err != nil {
				return err
			}
		}
		return nil
	default:
		if len(items) == 0 {
			_, err := fmt.Fprintln(w, "No items.")
			return err
		}
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		_, err := fmt.Fprintln(tw, "STATUS\tTITLE\tTRAY\tCREATED")
		if err != nil {
			return err
		}
		for _, it := range items {
			tn := trayNames[it.TrayID]
			if tn == "" {
				tn = it.TrayID
			}
			_, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
				it.Status, it.Title, tn, formatTrayLocalTime(it.CreatedAt))
			if err != nil {
				return err
			}
		}
		return tw.Flush()
	}
}
