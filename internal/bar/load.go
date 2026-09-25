package bar

import (
	"strings"

	"github.com/tjdsneto/tray-cli/internal/domain"
	"github.com/tjdsneto/tray-cli/internal/localtray"
)

func ItemsFromLocal(rows []localtray.ItemWithTray) []Item {
	out := make([]Item, 0, len(rows))
	for _, r := range rows {
		name := strings.TrimSpace(r.Tray.Name)
		if name == "" {
			name = localtray.DisplayLabel(r.Tray.ID)
		}
		out = append(out, Item{
			Scope: ScopeLocal, TrayRef: LocalTrayRef(r.Tray), TrayName: name,
			ItemID: r.Item.ID, Title: r.Item.Title, CreatedAt: r.Item.CreatedAt,
		})
	}
	return out
}

func ItemsFromRemote(items []domain.Item, trayNames map[string]string) []Item {
	out := make([]Item, 0, len(items))
	for _, it := range items {
		name := strings.TrimSpace(trayNames[it.TrayID])
		if name == "" {
			name = it.TrayID
		}
		out = append(out, Item{
			Scope: ScopeRemote, TrayRef: name, TrayName: name,
			ItemID: it.ID, Title: it.Title, CreatedAt: it.CreatedAt,
		})
	}
	return out
}
