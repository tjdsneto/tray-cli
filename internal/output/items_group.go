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

func sortItemsInTrayOrder(items []domain.Item) {
	sort.SliceStable(items, func(i, j int) bool {
		a, b := items[i], items[j]
		if a.TrayID != b.TrayID {
			return a.TrayID < b.TrayID
		}
		if a.SortOrder != b.SortOrder {
			return a.SortOrder < b.SortOrder
		}
		return a.CreatedAt.Before(b.CreatedAt)
	})
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
