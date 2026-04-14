// Package trayref resolves tray names, ids, and remote aliases without I/O beyond injected list calls.
package trayref

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/tjdsneto/tray-cli/internal/domain"
)

// FindTraysByNameFold returns trays whose name matches (case-insensitive, trimmed).
func FindTraysByNameFold(trays []domain.Tray, name string) []domain.Tray {
	want := strings.TrimSpace(name)
	if want == "" {
		return nil
	}
	var out []domain.Tray
	for i := range trays {
		if strings.EqualFold(strings.TrimSpace(trays[i].Name), want) {
			out = append(out, trays[i])
		}
	}
	return out
}

// PickTrayOrError returns the single matching tray or an error if none / ambiguous.
func PickTrayOrError(matches []domain.Tray, name string) (domain.Tray, error) {
	switch len(matches) {
	case 0:
		return domain.Tray{}, fmt.Errorf("no tray named %q — run `tray ls` to see names you can use", strings.TrimSpace(name))
	case 1:
		return matches[0], nil
	default:
		return domain.Tray{}, fmt.Errorf("multiple trays match %q — rename one in the dashboard or use a more specific name", strings.TrimSpace(name))
	}
}

// TrayNameMap builds tray id → name for display (ListMine output).
func TrayNameMap(trays []domain.Tray) map[string]string {
	m := make(map[string]string, len(trays))
	for i := range trays {
		m[trays[i].ID] = trays[i].Name
	}
	return m
}

// OverlayTrayAliases returns a copy of base with tray display names replaced when remotes.json maps an alias to that tray id.
// aliases is the raw map from remotes (alias label → tray uuid). If several aliases point at the same tray, the lexicographically smallest alias is used.
func OverlayTrayAliases(base map[string]string, aliases map[string]string) map[string]string {
	if len(aliases) == 0 {
		return base
	}
	byTray := make(map[string][]string)
	for alias, tid := range aliases {
		a := strings.TrimSpace(alias)
		t := strings.TrimSpace(tid)
		if a == "" || t == "" {
			continue
		}
		byTray[t] = append(byTray[t], a)
	}
	if len(byTray) == 0 {
		return base
	}
	for t := range byTray {
		sort.Strings(byTray[t])
	}
	out := make(map[string]string, len(base)+len(byTray))
	for k, v := range base {
		out[k] = v
	}
	for tid, as := range byTray {
		if len(as) == 0 {
			continue
		}
		out[tid] = as[0]
	}
	return out
}

// TrayIDFromRef resolves a tray reference using only in-memory data (pure).
// aliases maps lowercased alias → tray uuid (optional; used by remote).
func TrayIDFromRef(ref string, aliases map[string]string, trays []domain.Tray) (string, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", fmt.Errorf("empty tray reference")
	}
	if aliases != nil {
		if id, ok := aliases[strings.ToLower(ref)]; ok && strings.TrimSpace(id) != "" {
			return strings.TrimSpace(id), nil
		}
	}
	for i := range trays {
		if trays[i].ID == ref {
			return ref, nil
		}
	}
	t, err := PickTrayOrError(FindTraysByNameFold(trays, ref), ref)
	if err != nil {
		return "", err
	}
	return t.ID, nil
}

// ResolveTrayRef resolves a tray id or name to a tray id. Remote aliases hit without calling the API.
func ResolveTrayRef(ctx context.Context, svcs domain.Services, sess domain.Session, ref string, aliases map[string]string) (string, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", fmt.Errorf("empty tray reference")
	}
	if aliases != nil {
		if id, ok := aliases[strings.ToLower(ref)]; ok && strings.TrimSpace(id) != "" {
			return strings.TrimSpace(id), nil
		}
	}
	trays, err := svcs.Trays.ListMine(ctx, sess)
	if err != nil {
		return "", err
	}
	return TrayIDFromRef(ref, nil, trays)
}

// TrayByID returns the tray with the given id from a ListMine slice.
func TrayByID(trays []domain.Tray, id string) (domain.Tray, bool) {
	for i := range trays {
		if trays[i].ID == id {
			return trays[i], true
		}
	}
	return domain.Tray{}, false
}
