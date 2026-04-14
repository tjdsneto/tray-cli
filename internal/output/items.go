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
	colOrd    = 4
	colStatus = 12
	colTitle  = 42
	colTray   = 14
	colBy     = 24 // name or email when profiles exist; otherwise short id
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
		if len(items) == 0 {
			_, err := fmt.Fprintln(w, "_No items._")
			return err
		}
		_, err := fmt.Fprintf(w, "| %s | %s | %s | %s | %s | %s |\n", "#", "Status", "Title", "Tray", "By", "Created")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(w, "| %s | %s | %s | %s | %s | %s |\n", "---", "---", "---", "---", "---", "---")
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
			by := strings.ReplaceAll(FormatSourceUser(it.SourceUserID, currentUserID, displayByID), "|", "\\|")
			_, err := fmt.Fprintf(w, "| %d | %s | %s | %s | %s | %s |\n",
				it.SortOrder,
				strings.ReplaceAll(it.Status, "|", "\\|"),
				ttl,
				tn,
				by,
				strings.ReplaceAll(HumanizeTimeAgo(it.CreatedAt, now), "|", "\\|"))
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
		color := ColorEnabled(w)
		sep := "  "
		_, err := fmt.Fprintf(w, "%s%s%s%s%s%s%s%s%s%s%s\n",
			padPlain("#", colOrd), sep,
			padPlain("STATUS", colStatus), sep,
			padPlain("TITLE", colTitle), sep,
			padPlain("TRAY", colTray), sep,
			padPlain("BY", colBy), sep,
			"CREATED")
		if err != nil {
			return err
		}
		for _, it := range items {
			tn := trayNames[it.TrayID]
			if tn == "" {
				tn = it.TrayID
			}
			statusCell := FormatStatusANSI(it.Status, color, colStatus)
			titleCell := padPlain(truncateRunesPlain(it.Title, colTitle), colTitle)
			trayCell := padPlain(truncateRunesPlain(tn, colTray), colTray)
			byCell := padPlain(truncateRunesPlain(FormatSourceUser(it.SourceUserID, currentUserID, displayByID), colBy), colBy)
			when := HumanizeTimeAgo(it.CreatedAt, now)
			ordCell := padPlain(fmt.Sprintf("%d", it.SortOrder), colOrd)
			_, err := fmt.Fprintf(w, "%s%s%s%s%s%s%s%s%s%s%s\n",
				ordCell, sep,
				statusCell, sep, titleCell, sep, trayCell, sep, byCell, sep, when)
			if err != nil {
				return err
			}
		}
		return nil
	}
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
