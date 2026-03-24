package commands

import (
	"context"
	"strings"

	"github.com/tjdsneto/tray-cli/internal/domain"
)

// profileDisplayMap resolves source user ids to display labels (name or email).
// On error or nil Profiles service, returns an empty map (callers fall back to short id).
func profileDisplayMap(ctx context.Context, sess domain.Session, svc domain.Services, userIDs []string) map[string]string {
	if svc.Profiles == nil {
		return map[string]string{}
	}
	m, err := svc.Profiles.LookupDisplay(ctx, sess, userIDs)
	if err != nil || m == nil {
		return map[string]string{}
	}
	return m
}

func sourceUserIDsFromItems(items []domain.Item) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, it := range items {
		id := strings.TrimSpace(it.SourceUserID)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}
