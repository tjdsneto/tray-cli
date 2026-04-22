package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/tjdsneto/tray-cli/internal/cli/itemref"
	"github.com/tjdsneto/tray-cli/internal/domain"
)

// itemIDPool is the set of items considered when resolving a non-canonical id prefix.
// Full UUIDs skip this list entirely.
type itemIDPool int

const (
	// poolReviewPending is pending lines on trays you own (same scope as tray review; accept and decline prefixes).
	poolReviewPending itemIDPool = iota
	// poolPendingAcceptedOwned is pending plus accepted on trays you own — short prefix match for complete/snooze without scanning archived history.
	poolPendingAcceptedOwned
	// poolArchiveCandidates is pending, accepted, and snoozed on owned trays.
	poolArchiveCandidates
	// poolReorderOwned is every item on trays you own (any status), for item up/down.
	poolReorderOwned
	// poolRemoveCandidates is owned-tray items plus your contributed (outbox) lines.
	poolRemoveCandidates
)

func (p itemIDPool) describe() string {
	switch p {
	case poolReviewPending:
		return "pending items on trays you own"
	case poolPendingAcceptedOwned:
		return "pending and accepted items on trays you own"
	case poolArchiveCandidates:
		return "pending, accepted, and snoozed items on trays you own"
	case poolReorderOwned:
		return "items on trays you own"
	case poolRemoveCandidates:
		return "items on trays you own plus lines you filed on others' trays"
	default:
		return "items"
	}
}

// resolveItemIDArg accepts a canonical UUID, a 32-digit hex UUID without hyphens, or a unique hex prefix
// (at least itemref.MinPrefixHex digits) among the items defined by pool.
func resolveItemIDArg(ctx context.Context, svcs domain.Services, sess domain.Session, raw string, pool itemIDPool) (string, error) {
	full, prefix, err := itemref.ParseItemRef(raw)
	if err != nil {
		return "", err
	}
	if full != "" {
		return full, nil
	}
	items, err := listItemsForPool(ctx, svcs, sess, pool)
	if err != nil {
		return "", err
	}
	id, err := itemref.ResolveAmongItems(items, prefix)
	if err != nil {
		return "", fmt.Errorf("resolving id among %s: %w", pool.describe(), err)
	}
	return id, nil
}

func ownedTrayIDs(ctx context.Context, svcs domain.Services, sess domain.Session) ([]string, error) {
	owned, err := svcs.Trays.ListOwned(ctx, sess)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(owned))
	for _, t := range owned {
		if tid := strings.TrimSpace(t.ID); tid != "" {
			ids = append(ids, tid)
		}
	}
	return ids, nil
}

func listItemsForPool(ctx context.Context, svcs domain.Services, sess domain.Session, pool itemIDPool) ([]domain.Item, error) {
	switch pool {
	case poolReviewPending:
		return listOwnedByStatus(ctx, svcs, sess, "pending")
	case poolPendingAcceptedOwned:
		return listOwnedByStatuses(ctx, svcs, sess, []string{"pending", "accepted"})
	case poolArchiveCandidates:
		return listOwnedByStatuses(ctx, svcs, sess, []string{"pending", "accepted", "snoozed"})
	case poolReorderOwned:
		return listOwnedAll(ctx, svcs, sess)
	case poolRemoveCandidates:
		return listOwnedAllPlusOutbox(ctx, svcs, sess)
	default:
		return nil, fmt.Errorf("internal: unknown item id pool %d", pool)
	}
}

func listOwnedByStatus(ctx context.Context, svcs domain.Services, sess domain.Session, status string) ([]domain.Item, error) {
	ids, err := ownedTrayIDs(ctx, svcs, sess)
	if err != nil || len(ids) == 0 {
		return nil, err
	}
	return svcs.Items.List(ctx, sess, domain.ListItemsQuery{Status: status, TrayIDIn: ids})
}

func listOwnedByStatuses(ctx context.Context, svcs domain.Services, sess domain.Session, statuses []string) ([]domain.Item, error) {
	ids, err := ownedTrayIDs(ctx, svcs, sess)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, nil
	}
	var merged []domain.Item
	seen := make(map[string]struct{})
	for _, st := range statuses {
		chunk, err := svcs.Items.List(ctx, sess, domain.ListItemsQuery{Status: st, TrayIDIn: ids})
		if err != nil {
			return nil, err
		}
		for _, it := range chunk {
			if _, ok := seen[it.ID]; ok {
				continue
			}
			seen[it.ID] = struct{}{}
			merged = append(merged, it)
		}
	}
	return merged, nil
}

func listOwnedAll(ctx context.Context, svcs domain.Services, sess domain.Session) ([]domain.Item, error) {
	ids, err := ownedTrayIDs(ctx, svcs, sess)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, nil
	}
	return svcs.Items.List(ctx, sess, domain.ListItemsQuery{TrayIDIn: ids})
}

func listOwnedAllPlusOutbox(ctx context.Context, svcs domain.Services, sess domain.Session) ([]domain.Item, error) {
	owned, err := listOwnedAll(ctx, svcs, sess)
	if err != nil {
		return nil, err
	}
	byID := make(map[string]domain.Item)
	for _, it := range owned {
		byID[it.ID] = it
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
