package domain

import "context"

// ProfileService resolves user id → display string (name or email) for CLI output.
type ProfileService interface {
	// LookupDisplay returns a map of user id → label. Missing ids are omitted.
	// On backend error (e.g. profiles table not migrated yet), returns an empty map.
	LookupDisplay(ctx context.Context, sess Session, userIDs []string) (map[string]string, error)
}
