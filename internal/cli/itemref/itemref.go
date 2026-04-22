// Package itemref parses tray item id arguments: full UUIDs, 32-char hex UUIDs, or unique hex prefixes.
package itemref

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/tjdsneto/tray-cli/internal/domain"
)

var fullUUIDRE = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// MinPrefixHex is the minimum number of hex digits (ignoring hyphens) accepted for prefix resolution.
const MinPrefixHex = 8

// ParseItemRef trims raw and returns either a canonical lowercase UUID (full non-empty, prefix empty)
// or a lowercase hex string without hyphens for prefix matching (full empty, prefix non-empty).
func ParseItemRef(raw string) (full string, prefixHex string, err error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", "", fmt.Errorf("empty item id")
	}
	if fullUUIDRE.MatchString(s) {
		return strings.ToLower(s), "", nil
	}
	compact := strings.ReplaceAll(strings.ToLower(s), "-", "")
	if compact == "" {
		return "", "", fmt.Errorf("empty item id")
	}
	for _, r := range compact {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
			return "", "", fmt.Errorf("item id must be a UUID or a hexadecimal prefix (hyphens optional)")
		}
	}
	if len(compact) == 32 {
		return formatUUID32(compact), "", nil
	}
	if len(compact) < MinPrefixHex {
		return "", "", fmt.Errorf("item id prefix must be at least %d hex digits to avoid collisions", MinPrefixHex)
	}
	if len(compact) > 32 {
		return "", "", fmt.Errorf("item id has too many hex digits")
	}
	return "", compact, nil
}

// ResolveAmongItems returns the single item id whose UUID (ignoring hyphens) starts with compactPrefix.
// compactPrefix must be lower-case hex, no hyphens.
func ResolveAmongItems(items []domain.Item, compactPrefix string) (string, error) {
	if compactPrefix == "" {
		return "", fmt.Errorf("internal: empty prefix")
	}
	var hits []string
	for _, it := range items {
		id := strings.TrimSpace(it.ID)
		if id == "" {
			continue
		}
		compactID := strings.ReplaceAll(strings.ToLower(id), "-", "")
		if strings.HasPrefix(compactID, compactPrefix) {
			hits = append(hits, id)
		}
	}
	if len(hits) == 0 {
		return "", fmt.Errorf("no item matches id prefix %q", compactPrefix)
	}
	if len(hits) > 1 {
		sort.Strings(hits)
		show := hits
		if len(show) > 5 {
			show = show[:5]
		}
		return "", fmt.Errorf("ambiguous item id prefix %q — matches %d items (e.g. %s); use a longer prefix or full id", compactPrefix, len(hits), strings.Join(show, ", "))
	}
	return hits[0], nil
}

func formatUUID32(lower32 string) string {
	if len(lower32) != 32 {
		return lower32
	}
	return lower32[0:8] + "-" + lower32[8:12] + "-" + lower32[12:16] + "-" + lower32[16:20] + "-" + lower32[20:32]
}
