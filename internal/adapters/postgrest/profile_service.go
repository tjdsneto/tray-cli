package postgrest

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/tjdsneto/tray-cli/internal/adapters/postgrest/pghttp"
	"github.com/tjdsneto/tray-cli/internal/config"
	"github.com/tjdsneto/tray-cli/internal/domain"
)

type profileService struct {
	pg *pghttp.Client
}

func newProfileService(pg *pghttp.Client) *profileService {
	return &profileService{pg: pg}
}

var _ domain.ProfileService = (*profileService)(nil)

type profileRow struct {
	ID       string  `json:"id"`
	Email    string  `json:"email"`
	FullName *string `json:"full_name"`
}

func (s *profileService) LookupDisplay(ctx context.Context, sess domain.Session, userIDs []string) (map[string]string, error) {
	seen := map[string]struct{}{}
	var ids []string
	for _, id := range userIDs {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return map[string]string{}, nil
	}
	path := profilesListPath(ids)
	var rows []profileRow
	if err := s.pg.DoJSON(ctx, sess.AccessToken, http.MethodGet, path, nil, &rows, nil); err != nil {
		if config.Debug() {
			log.Printf("tray: profiles lookup failed (By column falls back to short id): %v", err)
		}
		// Table may not exist until migration is applied, or request may be rejected (e.g. bad filter).
		return map[string]string{}, nil
	}
	out := make(map[string]string, len(rows))
	for _, r := range rows {
		id := strings.TrimSpace(r.ID)
		if id == "" {
			continue
		}
		if lab := profileDisplayLabel(r); lab != "" {
			out[id] = lab
		}
	}
	return out, nil
}

func profileDisplayLabel(r profileRow) string {
	if r.FullName != nil {
		if t := strings.TrimSpace(*r.FullName); t != "" {
			return t
		}
	}
	if t := strings.TrimSpace(r.Email); t != "" {
		return t
	}
	return ""
}

func profilesListPath(ids []string) string {
	v := url.Values{}
	v.Set("select", "id,email,full_name")
	// PostgREST parses `in.(...)` tokens; UUIDs contain hyphens that must be quoted or they are
	// mis-read as minus operators. See https://postgrest.org/en/stable/references/api/tables_views.html#operators
	var parts []string
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		id = strings.ReplaceAll(id, `"`, "")
		parts = append(parts, `"`+id+`"`)
	}
	v.Set("id", "in.("+strings.Join(parts, ",")+")")
	return "/rest/v1/profiles?" + v.Encode()
}
