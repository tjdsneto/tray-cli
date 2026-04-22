package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/tjdsneto/tray-cli/internal/domain"
	"golang.org/x/term"
)

func timeColumnHeader(sectionStatus string) string {
	switch strings.ToLower(strings.TrimSpace(sectionStatus)) {
	case "completed":
		return "COMPLETED ON"
	default:
		return "ADDED ON"
	}
}

func timeColumnHeaderMarkdown(sectionStatus string) string {
	switch strings.ToLower(strings.TrimSpace(sectionStatus)) {
	case "completed":
		return "Completed on"
	default:
		return "Added on"
	}
}

// itemTimeDisplayForSection picks the timestamp shown in the last column: completion time for completed rows when set.
func itemTimeDisplayForSection(it domain.Item, sectionStatus string, now time.Time) string {
	if strings.EqualFold(strings.TrimSpace(sectionStatus), "completed") && it.CompletedAt != nil {
		return HumanizeTimeAgo(*it.CompletedAt, now)
	}
	return HumanizeTimeAgo(it.CreatedAt, now)
}

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
		sortItemsForDisplay(chunk, trayNames)
		if !first {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}
		first = false
		if _, err := fmt.Fprintf(w, "### %s\n\n", sectionTitleForStatus(st)); err != nil {
			return err
		}
		groups := groupConsecutiveByTrayID(chunk)
		firstGroup := true
		for _, group := range groups {
			if !firstGroup {
				if _, err := fmt.Fprintln(w); err != nil {
					return err
				}
			}
			firstGroup = false
			tn := itemTrayDisplayName(group[0], trayNames)
			heading := markdownTrayHeading(tn)
			if _, err := fmt.Fprintf(w, "#### %s\n\n", heading); err != nil {
				return err
			}
			_, err := fmt.Fprintf(w, "| %s | %s | %s | %s | %s |\n", "#", "id", "Title", "By", timeColumnHeaderMarkdown(st))
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(w, "| %s | %s | %s | %s | %s |\n", "---", "---", "---", "---", "---")
			if err != nil {
				return err
			}
			for i, it := range group {
				ttl := strings.ReplaceAll(it.Title, "|", "\\|")
				by := strings.ReplaceAll(FormatSourceUser(it.SourceUserID, currentUserID, displayByID), "|", "\\|")
				when := itemTimeDisplayForSection(it, st, now)
				idCell := "`" + strings.ReplaceAll(it.ID, "`", "") + "`"
				if _, err := fmt.Fprintf(w, "| %d | %s | %s | %s | %s |\n",
					i+1,
					idCell,
					ttl,
					by,
					strings.ReplaceAll(when, "|", "\\|")); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func itemTrayDisplayName(it domain.Item, trayNames map[string]string) string {
	if trayNames == nil {
		trayNames = map[string]string{}
	}
	tn := trayNames[it.TrayID]
	if tn == "" {
		return it.TrayID
	}
	return tn
}

func markdownTrayHeading(name string) string {
	s := strings.TrimSpace(strings.ReplaceAll(name, "\n", " "))
	s = strings.ReplaceAll(s, "#", "")
	return strings.TrimSpace(s)
}

func writeItemsTableGrouped(w io.Writer, items []domain.Item, trayNames map[string]string, currentUserID string, displayByID map[string]string, now time.Time) error {
	if len(items) == 0 {
		_, err := fmt.Fprintln(w, "No items.")
		return err
	}
	color := ColorEnabled(w)
	lineWidth := resolvedLineWidth(w)
	titleWrap := lineWidth - 2 // two-space indent for title continuation lines
	if titleWrap < 24 {
		titleWrap = 24
	}
	buckets := partitionItemsByStatus(items)
	keys := sectionKeysInDisplayOrder(buckets)
	first := true
	for _, st := range keys {
		chunk := buckets[st]
		sortItemsForDisplay(chunk, trayNames)
		if !first {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}
		first = false
		title := sectionTitleForStatus(st)
		if _, err := fmt.Fprintf(w, "%s\n", StatusSectionTitleANSI(st, title, color)); err != nil {
			return err
		}
		groups := groupConsecutiveByTrayID(chunk)
		for gi, group := range groups {
			if gi > 0 {
				if _, err := fmt.Fprintln(w); err != nil {
					return err
				}
			}
			tn := itemTrayDisplayName(group[0], trayNames)
			if _, err := fmt.Fprintf(w, "%s\n", TrayGroupTitleANSI(tn, color)); err != nil {
				return err
			}
			for ii, it := range group {
				if ii > 0 {
					if _, err := fmt.Fprintln(w); err != nil {
						return err
					}
				}
				by := FormatSourceUser(it.SourceUserID, currentUserID, displayByID)
				when := itemTimeDisplayForSection(it, st, now)
				head := by + " · " + when
				suffixPlain := " · " + strings.TrimSpace(it.ID)
				meta := formatItemMetaHuman(ii+1, head, suffixPlain, lineWidth, color)
				if _, err := fmt.Fprintln(w, meta); err != nil {
					return err
				}
				for _, line := range wrapPlainTitle(it.Title, titleWrap) {
					if _, err := fmt.Fprintf(w, "  %s\n", line); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// formatItemMetaHuman builds one meta line: display item number (1-based within this tray block),
// contributor and time, then uuid. sort_order is not shown; it only affects sort order upstream.
// On a TTY: bold cyan line number, dim "who · when", yellow id (tray name is colored separately).
func formatItemMetaHuman(displayN int, headPlain, suffixPlain string, lineWidth int, color bool) string {
	prefix := fmt.Sprintf("%3d  ", displayN)
	s := fitItemMetaLine(prefix, headPlain, suffixPlain, lineWidth)
	if !color {
		return s
	}
	if strings.HasSuffix(s, suffixPlain) {
		headPart := s[len(prefix) : len(s)-len(suffixPlain)]
		numColored := ansiBold + ansiCyan + fmt.Sprintf("%3d", displayN) + ansiReset + "  "
		dimHead := ansiDim + headPart + ansiReset
		idColored := ansiYellow + suffixPlain + ansiReset
		return numColored + dimHead + idColored
	}
	return ansiDim + s + ansiReset
}

// fitItemMetaLine joins prefix (may contain ANSI), plain head, and suffix (may contain ANSI).
// If the visible width exceeds lineWidth, head is truncated with an ellipsis so the full suffix stays.
func fitItemMetaLine(prefix, head, suffix string, lineWidth int) string {
	if lineWidth < 48 {
		lineWidth = 48
	}
	pLen := utf8.RuneCountInString(StripANSI(prefix))
	sLen := utf8.RuneCountInString(StripANSI(suffix))
	hLen := utf8.RuneCountInString(head)
	if pLen+hLen+sLen <= lineWidth {
		return prefix + head + suffix
	}
	maxHead := lineWidth - pLen - sLen
	if maxHead < 1 {
		maxHead = 1
	}
	return prefix + truncateRunesPlain(head, maxHead) + suffix
}

// resolvedLineWidth returns stdout width when w is a TTY *os.File, else a default suitable for pipes and tests.
func resolvedLineWidth(w io.Writer) int {
	const defaultCols = 100
	const minCols = 60
	f, ok := w.(*os.File)
	if !ok || !term.IsTerminal(int(f.Fd())) {
		return defaultCols
	}
	cols, _, err := term.GetSize(int(f.Fd()))
	if err != nil || cols < minCols {
		return defaultCols
	}
	return cols
}

// wrapPlainTitle folds a tray item title into lines of at most width runes, preferring word boundaries.
// Newlines in s are treated as spaces.
func wrapPlainTitle(s string, width int) []string {
	s = strings.TrimSpace(strings.ReplaceAll(s, "\n", " "))
	if s == "" {
		return nil
	}
	if width < 8 {
		width = 8
	}
	rs := []rune(s)
	var out []string
	for len(rs) > 0 {
		rs = trimSpaceRunes(rs)
		if len(rs) == 0 {
			break
		}
		if len(rs) <= width {
			out = append(out, string(rs))
			break
		}
		line, rest := splitGreedyWordRunes(rs, width)
		line = trimSpaceRunes(line)
		if len(line) == 0 {
			// Degenerate: leading spaces only in window, or all-space prefix — hard-break.
			out = append(out, string(rs[:width]))
			rs = rs[width:]
			continue
		}
		out = append(out, string(line))
		rs = rest
	}
	return out
}

func trimSpaceRunes(rs []rune) []rune {
	i, j := 0, len(rs)
	for i < j && rs[i] == ' ' {
		i++
	}
	for j > i && rs[j-1] == ' ' {
		j--
	}
	return rs[i:j]
}

func splitGreedyWordRunes(rs []rune, width int) (line, rest []rune) {
	if len(rs) <= width {
		return rs, nil
	}
	prefix := rs[:width]
	for i := len(prefix) - 1; i > 0; i-- {
		if prefix[i] == ' ' {
			return prefix[:i], rs[i+1:]
		}
	}
	return prefix, rs[width:]
}

func itemTimeJSON(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.UTC().Format(time.RFC3339)
	return &s
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
