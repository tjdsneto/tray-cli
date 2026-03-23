package postgrest

import (
	"net/http"

	"github.com/tjdsneto/tray-cli/internal/adapters/postgrest/pghttp"
	"github.com/tjdsneto/tray-cli/internal/domain"
	supabasehttp "github.com/tjdsneto/tray-cli/internal/supabase"
)

// NewServices wires PostgREST-backed domain services around a Supabase HTTP client.
func NewServices(c *supabasehttp.Client) domain.Services {
	pg := pghttp.New(c)
	return domain.Services{
		Trays: newTrayService(pg),
		Items: newItemService(pg),
	}
}

// Dial builds the HTTP client and returns domain.Services in one step (typical CLI wiring).
func Dial(rawURL, anonKey string, httpClient *http.Client) (domain.Services, error) {
	c, err := supabasehttp.NewClient(rawURL, anonKey, httpClient)
	if err != nil {
		return domain.Services{}, err
	}
	return NewServices(c), nil
}
