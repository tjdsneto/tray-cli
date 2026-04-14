package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/tjdsneto/tray-cli/internal/domain"
)

const (
	colOrd   = 4  // manual order (sort_order) within each tray
	colTitle = 52 // wider: STATUS column removed from grouped list
	colTray  = 14
	colBy    = 24 // name or email when profiles exist; otherwise short id
	colAdded = 16 // "ADDED ON" / relative time
)

// WriteItems renders items; trayNames maps tray_id → display name (optional).
// currentUserID is used to show "you" for SourceUserID when it matches (can be empty).
// displayByID maps source_user_id → label from profiles (name or email); nil or empty entries fall back to a short id.
func WriteItems(w io.Writer, items []domain.Item, trayNames map[string]string, currentUserID string, displayByID map[string]string, f Format) error {
	if trayNames == nil {
		trayNames = map[string]string{}
	}
	now := time.Now()
	switch f {
	case FormatJSON:
		type row struct {
			ID               string  `json:"id"`
			TrayID           string  `json:"tray_id"`
			TrayName         string  `json:"tray_name,omitempty"`
			SortOrder        int     `json:"sort_order"`
			Title            string  `json:"title"`
			Status           string  `json:"status"`
			DueDate          *string `json:"due_date,omitempty"`
			CreatedAt        string  `json:"created_at"`
			CreatedAgo       string  `json:"created_ago"`
			SourceUserID     string  `json:"source_user_id"`
			SourceUserLabel  string  `json:"source_user_label"`
			AcceptedAt       *string `json:"accepted_at,omitempty"`
			DeclinedAt       *string `json:"declined_at,omitempty"`
			CompletedAt      *string `json:"completed_at,omitempty"`
			ArchivedAt       *string `json:"archived_at,omitempty"`
			SnoozedAt        *string `json:"snoozed_at,omitempty"`
		}
		out := make([]row, 0, len(items))
		for _, it := range items {
			out = append(out, row{
				ID:              it.ID,
				TrayID:          it.TrayID,
				TrayName:        trayNames[it.TrayID],
				SortOrder:       it.SortOrder,
				Title:           it.Title,
				Status:          it.Status,
				DueDate:         it.DueDate,
				CreatedAt:       it.CreatedAt.UTC().Format(time.RFC3339),
				CreatedAgo:      HumanizeTimeAgo(it.CreatedAt, now),
				SourceUserID:    it.SourceUserID,
				SourceUserLabel: FormatSourceUser(it.SourceUserID, currentUserID, displayByID),
				AcceptedAt:      itemTimeJSON(it.AcceptedAt),
				DeclinedAt:      itemTimeJSON(it.DeclinedAt),
				CompletedAt:     itemTimeJSON(it.CompletedAt),
				ArchivedAt:      itemTimeJSON(it.ArchivedAt),
				SnoozedAt:       itemTimeJSON(it.SnoozedAt),
			})
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(out)
	case FormatMarkdown:
		return writeItemsMarkdownGrouped(w, items, trayNames, currentUserID, displayByID, now)
	default:
		return writeItemsTableGrouped(w, items, trayNames, currentUserID, displayByID, now)
	}
}

func writeItemsMarkdownGrouped(w io.Writer, items []domain.Item, trayNames map[string]string, currentUserID string, displayByID map[string]string, now time.Time) error {
	if len(items) == 0 {
		_, err := fmt.Fprintln(w, "_No items._")
		return err
	}
	buckets := partitionItemsByStatus(items)
	keys := sectionKeysInDisplayOrder(buckets)
	first := true
	for _, st := range keys {
		chunk := buckets[st]
		sortItemsInTrayOrder(chunk)
		if !first {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}
		first = false
		if _, err := fmt.Fprintf(w, "### %s\n\n", sectionTitleForStatus(st)); err != nil {
			return err
		}
		_, err := fmt.Fprintf(w, "| %s | %s | %s | %s | %s |\n", "ORD", "Title", "Tray", "By", "Added on")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(w, "| %s | %s | %s | %s | %s |\n", "---", "---", "---", "---", "---")
		if err != nil {
			return err
		}
		for _, it := range chunk {
			tn := trayNames[it.TrayID]
			if tn == "" {
				tn = it.TrayID
			}
			tn = strings.ReplaceAll(tn, "|", "\\|")
			ttl := strings.ReplaceAll(it.Title, "|", "\\|")
			by := strings.ReplaceAll(FormatSourceUser(it.SourceUserID, currentUserID, displayByID), "|", "\\|")
			added := truncateRunesPlain(HumanizeTimeAgo(it.CreatedAt, now), 24)
			if _, err := fmt.Fprintf(w, "| %d | %s | %s | %s | %s |\n",
				it.SortOrder,
				ttl,
				tn,
				by,
				strings.ReplaceAll(added, "|", "\\|")); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeItemsTableGrouped(w io.Writer, items []domain.Item, trayNames map[string]string, currentUserID string, displayByID map[string]string, now time.Time) error {
	if len(items) == 0 {
		_, err := fmt.Fprintln(w, "No items.")
		return err
	}
	color := ColorEnabled(w)
	sep := "  "
	buckets := partitionItemsByStatus(items)
	keys := sectionKeysInDisplayOrder(buckets)
	first := true
	for _, st := range keys {
		chunk := buckets[st]
		sortItemsInTrayOrder(chunk)
		if !first {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}
		first = false
		title := sectionTitleForStatus(st)
		if color {
			if _, err := fmt.Fprintf(w, "\x1b[1m%s\x1b[0m\n", title); err != nil {
				return err
			}
		} else {
			if _, err := fmt.Fprintf(w, "%s\n", title); err != nil {
				return err
			}
		}
		_, err := fmt.Fprintf(w, "%s%s%s%s%s%s%s%s%s\n",
			padPlain("ORD", colOrd), sep,
			padPlain("TITLE", colTitle), sep,
			padPlain("TRAY", colTray), sep,
			padPlain("BY", colBy), sep,
			padPlain("ADDED ON", colAdded))
		if err != nil {
			return err
		}
		for _, it := range chunk {
			tn := trayNames[it.TrayID]
			if tn == "" {
				tn = it.TrayID
			}
			titleCell := padPlain(truncateRunesPlain(it.Title, colTitle), colTitle)
			trayCell := padPlain(truncateRunesPlain(tn, colTray), colTray)
			byCell := padPlain(truncateRunesPlain(FormatSourceUser(it.SourceUserID, currentUserID, displayByID), colBy), colBy)
			when := padPlain(truncateRunesPlain(HumanizeTimeAgo(it.CreatedAt, now), colAdded), colAdded)
			ordCell := padPlain(fmt.Sprintf("%d", it.SortOrder), colOrd)
			_, err := fmt.Fprintf(w, "%s%s%s%s%s%s%s%s%s\n",
				ordCell, sep,
				titleCell, sep, trayCell, sep, byCell, sep, when)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func itemTimeJSON(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.UTC().Format(time.RFC3339)
	return &s
}

func padPlain(s string, width int) string {
	n := utf8.RuneCountInString(s)
	if n >= width {
		return s
	}
	return s + strings.Repeat(" ", width-n)
}

func truncateRunesPlain(s string, max int) string {
	if max <= 0 {
		return ""
	}
	rs := []rune(s)
	if len(rs) <= max {
		return s
	}
	if max == 1 {
		return "…"
	}
	return string(rs[:max-1]) + "…"
}
