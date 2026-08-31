package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/tjdsneto/tray-cli/internal/localtray"
)

// WriteLocalItems renders local tray items grouped by tray label.
func WriteLocalItems(w io.Writer, rows []localtray.ItemWithTray, f Format) error {
	switch f {
	case FormatJSON:
		type row struct {
			ID         string `json:"id"`
			TrayID     string `json:"tray_id"`
			TrayLabel  string `json:"tray_label"`
			Title      string `json:"title"`
			Status     string `json:"status"`
			CreatedAt  string `json:"created_at"`
			CreatedAgo string `json:"created_ago"`
		}
		now := time.Now()
		out := make([]row, 0, len(rows))
		for _, r := range rows {
			out = append(out, row{
				ID:         r.Item.ID,
				TrayID:     r.Item.TrayID,
				TrayLabel:  localtray.DisplayLabel(r.Item.TrayID),
				Title:      r.Item.Title,
				Status:     r.Item.Status,
				CreatedAt:  r.Item.CreatedAt.UTC().Format(time.RFC3339),
				CreatedAgo: HumanizeTimeAgo(r.Item.CreatedAt, now),
			})
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(out)
	case FormatMarkdown:
		if len(rows) == 0 {
			_, err := fmt.Fprint(w, "_No local items in scope._\n")
			return err
		}
		var cur string
		for _, r := range rows {
			label := localtray.DisplayLabel(r.Item.TrayID)
			if label != cur {
				if cur != "" {
					if _, err := fmt.Fprintln(w); err != nil {
						return err
					}
				}
				cur = label
				if _, err := fmt.Fprintf(w, "### %s\n\n", label); err != nil {
					return err
				}
			}
			if _, err := fmt.Fprintf(w, "- `%s` · %s · _%s_\n", shortID(r.Item.ID), r.Item.Title, HumanizeTimeAgo(r.Item.CreatedAt, time.Now())); err != nil {
				return err
			}
		}
		return nil
	default:
		if len(rows) == 0 {
			_, err := fmt.Fprintln(w, "No local items in scope.")
			return err
		}
		var cur string
		n := 0
		for _, r := range rows {
			label := localtray.DisplayLabel(r.Item.TrayID)
			if label != cur {
				cur = label
				n = 0
				if _, err := fmt.Fprintf(w, "\n%s\n", label); err != nil {
					return err
				}
			}
			n++
			if _, err := fmt.Fprintf(w, "  %d · %s · %s · %s\n", n, shortID(r.Item.ID), HumanizeTimeAgo(r.Item.CreatedAt, time.Now()), r.Item.Title); err != nil {
				return err
			}
		}
		return nil
	}
}

func shortID(id string) string {
	id = strings.TrimSpace(id)
	if len(id) <= 12 {
		return id
	}
	return id[:8] + "…"
}

// WriteLocalTrays renders local tray summaries for tray ls.
func WriteLocalTrays(w io.Writer, summaries []localtray.TraySummary, f Format) error {
	switch f {
	case FormatJSON:
		type row struct {
			ID         string `json:"id"`
			Kind       string `json:"kind"`
			Label      string `json:"label"`
			OpenItems  int    `json:"open_items"`
			TotalItems int    `json:"total_items"`
			CreatedAt  string `json:"created_at"`
		}
		out := make([]row, 0, len(summaries))
		for _, s := range summaries {
			out = append(out, row{
				ID:         s.Record.ID,
				Kind:       s.Record.Kind,
				Label:      localtray.DisplayLabel(s.Record.ID),
				OpenItems:  s.OpenCount,
				TotalItems: s.TotalCount,
				CreatedAt:  s.Record.CreatedAt.UTC().Format(time.RFC3339),
			})
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(out)
	case FormatMarkdown:
		if len(summaries) == 0 {
			_, err := fmt.Fprint(w, "_No local trays in scope._\n")
			return err
		}
		_, err := fmt.Fprintf(w, "| %s | %s | %s | %s |\n", "Tray", "Open", "Total", "Created")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(w, "| %s | %s | %s | %s |\n", "---", "---", "---", "---")
		if err != nil {
			return err
		}
		for _, s := range summaries {
			label := strings.ReplaceAll(localtray.DisplayLabel(s.Record.ID), "|", "\\|")
			_, err := fmt.Fprintf(w, "| %s | %d | %d | %s |\n", label, s.OpenCount, s.TotalCount, formatTrayLocalTime(s.Record.CreatedAt))
			if err != nil {
				return err
			}
		}
		return nil
	default:
		if len(summaries) == 0 {
			_, err := fmt.Fprintln(w, "No local trays in scope.")
			return err
		}
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		if _, err := fmt.Fprintln(tw, "TRAY\tOPEN\tTOTAL\tCREATED"); err != nil {
			return err
		}
		for _, s := range summaries {
			if _, err := fmt.Fprintf(tw, "%s\t%d\t%d\t%s\n", localtray.DisplayLabel(s.Record.ID), s.OpenCount, s.TotalCount, formatTrayLocalTime(s.Record.CreatedAt)); err != nil {
				return err
			}
		}
		return tw.Flush()
	}
}

// WriteLocalItemAdded prints confirmation after a local add.
func WriteLocalItemAdded(w io.Writer, item localtray.Item, f Format) error {
	label := localtray.DisplayLabel(item.TrayID)
	switch f {
	case FormatJSON:
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(map[string]string{
			"id":         item.ID,
			"tray_id":    item.TrayID,
			"tray_label": label,
			"title":      item.Title,
			"status":     item.Status,
		})
	case FormatMarkdown:
		_, err := fmt.Fprintf(w, "Added to **%s**: `%s` — %s\n", label, item.ID, item.Title)
		return err
	default:
		_, err := fmt.Fprintf(w, "Added to %s: %s — %s\n", label, item.ID, item.Title)
		return err
	}
}
