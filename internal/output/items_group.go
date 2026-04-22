package output

import (
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/tjdsneto/tray-cli/internal/domain"
)

// Primary display order for tray list: accepted first, then triage pipeline, then terminal states.
var listStatusSectionOrder = []string{
	"accepted",
	"pending",
	"snoozed",
	"declined",
	"completed",
	"archived",
}

func partitionItemsByStatus(items []domain.Item) map[string][]domain.Item {
	m := make(map[string][]domain.Item)
	for _, it := range items {
		st := strings.ToLower(strings.TrimSpace(it.Status))
		if st == "" {
			st = "unknown"
		}
		m[st] = append(m[st], it)
	}
	return m
}

// sectionKeysInDisplayOrder returns non-empty status keys: known order first, then any other statuses (sorted).
func sectionKeysInDisplayOrder(buckets map[string][]domain.Item) []string {
	var keys []string
	seen := make(map[string]struct{})
	for _, st := range listStatusSectionOrder {
		if len(buckets[st]) == 0 {
			continue
		}
		keys = append(keys, st)
		seen[st] = struct{}{}
	}
	var rest []string
	for st := range buckets {
		if len(buckets[st]) == 0 {
			continue
		}
		if _, ok := seen[st]; ok {
			continue
		}
		rest = append(rest, st)
	}
	sort.Strings(rest)
	return append(keys, rest...)
}

// sortItemsForDisplay sorts by tray display name (case-insensitive), then tray id, manual order, then created time.
func sortItemsForDisplay(items []domain.Item, trayNames map[string]string) {
	if trayNames == nil {
		trayNames = map[string]string{}
	}
	sort.SliceStable(items, func(i, j int) bool {
		a, b := items[i], items[j]
		ka, kb := trayDisplaySortKey(a.TrayID, trayNames), trayDisplaySortKey(b.TrayID, trayNames)
		if ka != kb {
			return ka < kb
		}
		if a.TrayID != b.TrayID {
			return a.TrayID < b.TrayID
		}
		if a.SortOrder != b.SortOrder {
			return a.SortOrder < b.SortOrder
		}
		return a.CreatedAt.Before(b.CreatedAt)
	})
}

func trayDisplaySortKey(trayID string, trayNames map[string]string) string {
	id := strings.TrimSpace(trayID)
	name := strings.TrimSpace(trayNames[id])
	if name == "" {
		name = id
	}
	return strings.ToLower(name) + "\x00" + id
}

// groupConsecutiveByTrayID splits items into runs with the same tray_id (slice must already be sorted by tray).
func groupConsecutiveByTrayID(items []domain.Item) [][]domain.Item {
	if len(items) == 0 {
		return nil
	}
	var groups [][]domain.Item
	var cur []domain.Item
	curTray := ""
	for _, it := range items {
		tid := strings.TrimSpace(it.TrayID)
		if len(cur) == 0 || tid != curTray {
			if len(cur) > 0 {
				groups = append(groups, cur)
			}
			cur = nil
			curTray = tid
		}
		cur = append(cur, it)
	}
	if len(cur) > 0 {
		groups = append(groups, cur)
	}
	return groups
}

func sectionTitleForStatus(st string) string {
	switch strings.ToLower(strings.TrimSpace(st)) {
	case "accepted":
		return "Accepted"
	case "pending":
		return "Pending"
	case "snoozed":
		return "Snoozed"
	case "declined":
		return "Declined"
	case "completed":
		return "Completed"
	case "archived":
		return "Archived"
	default:
		r, size := utf8.DecodeRuneInString(st)
		if size == 0 {
			return st
		}
		return string(unicode.ToUpper(r)) + st[size:]
	}
}
