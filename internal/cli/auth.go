package cli

import (
	"fmt"
	"net/http"

	"github.com/tjdsneto/tray-cli/internal/adapters/postgrest"
	"github.com/tjdsneto/tray-cli/internal/config"
	"github.com/tjdsneto/tray-cli/internal/credentials"
	"github.com/tjdsneto/tray-cli/internal/domain"
)

// requireAuth loads Supabase config, dials PostgREST, and returns a session from saved credentials.
func requireAuth() (domain.Services, domain.Session, error) {
	rawURL := config.SupabaseURL()
	anon := config.SupabaseAnonKey()
	if rawURL == "" || anon == "" {
		return domain.Services{}, domain.Session{}, fmt.Errorf("set %s and %s (environment) or build with ./run.sh or ./build.sh and a .env", config.EnvSupabaseURL, config.EnvSupabaseAnonKey)
	}
	svcs, err := postgrest.Dial(rawURL, anon, http.DefaultClient)
	if err != nil {
		return domain.Services{}, domain.Session{}, err
	}
	f, err := credentials.Load(ConfigDir())
	if err != nil {
		return domain.Services{}, domain.Session{}, err
	}
	return svcs, f.Session(), nil
}
