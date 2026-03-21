package postgrest

import (
	"net/http"
	"testing"

	"github.com/tjdsneto/tray-cli/internal/config"
	"github.com/stretchr/testify/require"
)

func TestHTTPAPIError_userFacing(t *testing.T) {
	t.Setenv(config.EnvDebug, "")

	t.Run("42P17", func(t *testing.T) {
		raw := []byte(`{"code":"42P17","message":"infinite recursion detected in policy for relation \"trays\""}`)
		err := httpAPIError("POST", "/rest/v1/trays", "500 Internal Server Error", http.StatusInternalServerError, raw)
		require.Error(t, err)
		require.Contains(t, err.Error(), "migration")
		require.NotContains(t, err.Error(), "infinite recursion")
	})

	t.Run("debug shows raw body", func(t *testing.T) {
		t.Setenv(config.EnvDebug, "1")
		raw := []byte(`{"code":"42P17","message":"secret"}`)
		err := httpAPIError("POST", "/rest/v1/trays", "500 Internal Server Error", http.StatusInternalServerError, raw)
		require.Error(t, err)
		require.Contains(t, err.Error(), "secret")
	})
}
