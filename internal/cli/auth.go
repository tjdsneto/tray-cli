package cli

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/tjdsneto/tray-cli/internal/adapters/postgrest"
	"github.com/tjdsneto/tray-cli/internal/auth"
	"github.com/tjdsneto/tray-cli/internal/cli/errs"
	"github.com/tjdsneto/tray-cli/internal/config"
	"github.com/tjdsneto/tray-cli/internal/credentials"
	"github.com/tjdsneto/tray-cli/internal/domain"
)

// requireAuth loads Supabase config, dials PostgREST, and returns a session from saved credentials.
// OAuth logins persist a refresh token; when the access JWT is expired or near expiry, it is refreshed automatically.
func requireAuth() (domain.Services, domain.Session, error) {
	rawURL := config.SupabaseURL()
	anon := config.SupabaseAnonKey()
	if rawURL == "" || anon == "" {
		return domain.Services{}, domain.Session{}, fmt.Errorf("%w", errs.MissingBackendConfig)
	}
	f, err := credentials.Load(ConfigDir())
	if err != nil {
		return domain.Services{}, domain.Session{}, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	f, err = auth.EnsureFreshCredentials(ctx, rawURL, anon, http.DefaultClient, ConfigDir(), f)
	if err != nil {
		return domain.Services{}, domain.Session{}, err
	}
	svcs, err := postgrest.Dial(rawURL, anon, http.DefaultClient)
	if err != nil {
		return domain.Services{}, domain.Session{}, err
	}
	return svcs, f.Session(), nil
}
