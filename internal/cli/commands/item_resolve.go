package commands

import (
	"context"
	"strings"

	"github.com/tjdsneto/tray-cli/internal/cli/itemref"
	"github.com/tjdsneto/tray-cli/internal/domain"
)

// resolveItemIDArg accepts a canonical UUID, a 32-digit hex UUID without hyphens, or a unique hex prefix
// (at least itemref.MinPrefixHex digits) among items on trays you own plus your contributed (outbox) items.
func resolveItemIDArg(ctx context.Context, svcs domain.Services, sess domain.Session, raw string) (string, error) {
	full, prefix, err := itemref.ParseItemRef(raw)
	if err != nil {
		return "", err
	}
	if full != "" {
		return full, nil
	}
	items, err := listResolvableItems(ctx, svcs, sess)
	if err != nil {
		return "", err
	}
	return itemref.ResolveAmongItems(items, prefix)
}

func listResolvableItems(ctx context.Context, svcs domain.Services, sess domain.Session) ([]domain.Item, error) {
	trays, err := svcs.Trays.ListMine(ctx, sess)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(trays))
	for _, t := range trays {
		if tid := strings.TrimSpace(t.ID); tid != "" {
			ids = append(ids, tid)
		}
	}
	byID := make(map[string]domain.Item)
	if len(ids) > 0 {
		owned, err := svcs.Items.List(ctx, sess, domain.ListItemsQuery{TrayIDIn: ids})
		if err != nil {
			return nil, err
		}
		for _, it := range owned {
			byID[it.ID] = it
		}
	}
	outbox, err := svcs.Items.ListOutbox(ctx, sess)
	if err != nil {
		return nil, err
	}
	for _, it := range outbox {
		byID[it.ID] = it
	}
	out := make([]domain.Item, 0, len(byID))
	for _, it := range byID {
		out = append(out, it)
	}
	return out, nil
}
