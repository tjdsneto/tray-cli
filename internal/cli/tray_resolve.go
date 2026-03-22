package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/tjdsneto/tray-cli/internal/domain"
)

// findTraysByNameFold returns trays whose name matches (case-insensitive, trimmed).
func findTraysByNameFold(trays []domain.Tray, name string) []domain.Tray {
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

func pickTrayOrError(matches []domain.Tray, name string) (domain.Tray, error) {
	switch len(matches) {
	case 0:
		return domain.Tray{}, fmt.Errorf("no tray named %q — run `tray ls` to see names you can use", strings.TrimSpace(name))
	case 1:
		return matches[0], nil
	default:
		return domain.Tray{}, fmt.Errorf("multiple trays match %q — rename one in the dashboard or use a more specific name", strings.TrimSpace(name))
	}
}

// trayNameMap builds tray id → name for display (ListMine output).
func trayNameMap(trays []domain.Tray) map[string]string {
	m := make(map[string]string, len(trays))
	for i := range trays {
		m[trays[i].ID] = trays[i].Name
	}
	return m
}

// resolveTrayRef resolves a tray id or name to a tray id. aliases maps lowercased alias → tray uuid (optional; used by remote).
func resolveTrayRef(ctx context.Context, svcs domain.Services, sess domain.Session, ref string, aliases map[string]string) (string, error) {
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
	for i := range trays {
		if trays[i].ID == ref {
			return ref, nil
		}
	}
	t, err := pickTrayOrError(findTraysByNameFold(trays, ref), ref)
	if err != nil {
		return "", err
	}
	return t.ID, nil
}
