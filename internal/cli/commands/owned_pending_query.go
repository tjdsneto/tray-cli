package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/tjdsneto/tray-cli/internal/cli/trayref"
	"github.com/tjdsneto/tray-cli/internal/domain"
)

// noOwnedTraysTrayFilter matches no real tray; used when the user owns zero trays so API queries return empty rows instead of all visible pending items.
const noOwnedTraysTrayFilter = "00000000-0000-0000-0000-000000000000"

func optionalTrayRefArg(args []string) string {
	if len(args) == 0 {
		return ""
	}
	return strings.TrimSpace(args[0])
}

// pendingItemsOnOwnedTraysQuery builds a ListItemsQuery for owner triage: pending items only on trays the current user owns.
func pendingItemsOnOwnedTraysQuery(ctx context.Context, svcs domain.Services, sess domain.Session, trayRef string, aliases map[string]string) (domain.ListItemsQuery, error) {
	q := domain.ListItemsQuery{Status: "pending"}
	owned, err := svcs.Trays.ListOwned(ctx, sess)
	if err != nil {
		return domain.ListItemsQuery{}, err
	}
	ownedIDs := make(map[string]struct{}, len(owned))
	for i := range owned {
		ownedIDs[strings.TrimSpace(owned[i].ID)] = struct{}{}
	}
	ref := strings.TrimSpace(trayRef)
	if ref != "" {
		tid, err := trayref.ResolveTrayRef(ctx, svcs, sess, ref, aliases)
		if err != nil {
			return domain.ListItemsQuery{}, err
		}
		tid = strings.TrimSpace(tid)
		if _, ok := ownedIDs[tid]; !ok {
			return domain.ListItemsQuery{}, fmt.Errorf("only trays you own — %q is not yours (items you added elsewhere: `tray contributed`; joined trays: `tray remote ls`)", ref)
		}
		q.TrayID = tid
		return q, nil
	}
	if len(owned) == 0 {
		q.TrayID = noOwnedTraysTrayFilter
		return q, nil
	}
	q.TrayIDIn = make([]string, 0, len(owned))
	for i := range owned {
		q.TrayIDIn = append(q.TrayIDIn, strings.TrimSpace(owned[i].ID))
	}
	return q, nil
}
