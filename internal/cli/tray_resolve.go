package cli

import (
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
