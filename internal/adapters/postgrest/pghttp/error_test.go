package pghttp

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tjdsneto/tray-cli/internal/config"
)

func TestAPIError_userFacing(t *testing.T) {
	t.Setenv(config.EnvDebug, "")

	t.Run("42P17", func(t *testing.T) {
		raw := []byte(`{"code":"42P17","message":"infinite recursion detected in policy for relation \"trays\""}`)
		err := apiError("POST", "/rest/v1/trays", "500 Internal Server Error", http.StatusInternalServerError, raw)
		require.Error(t, err)
		require.Contains(t, err.Error(), "migration")
		require.NotContains(t, err.Error(), "infinite recursion")
	})

	t.Run("debug shows raw body", func(t *testing.T) {
		t.Setenv(config.EnvDebug, "1")
		raw := []byte(`{"code":"42P17","message":"secret"}`)
		err := apiError("POST", "/rest/v1/trays", "500 Internal Server Error", http.StatusInternalServerError, raw)
		require.Error(t, err)
		require.Contains(t, err.Error(), "secret")
	})

	t.Run("duplicate tray name", func(t *testing.T) {
		t.Setenv(config.EnvDebug, "")
		raw := []byte(`{"code":"23505","message":"duplicate key value violates unique constraint trays_owner_name_unique"}`)
		err := apiError("POST", "/rest/v1/trays", "409 Conflict", http.StatusConflict, raw)
		require.Error(t, err)
		require.Contains(t, err.Error(), "already have a tray")
		require.Contains(t, err.Error(), "tray ls")
	})
}

func TestAPIError_statuses(t *testing.T) {
	t.Setenv(config.EnvDebug, "")
	t.Cleanup(func() { t.Setenv(config.EnvDebug, "") })

	cases := []struct {
		name   string
		status int
		raw    string
		substr string
	}{
		{"400 with body", http.StatusBadRequest, `{"message":"bad field"}`, "wasn't valid"},
		{"401", http.StatusUnauthorized, `{}`, "session expired"},
		{"403 with msg", http.StatusForbidden, `{"message":"RLS"}`, "not allowed"},
		{"403 empty", http.StatusForbidden, `{}`, "permission denied"},
		{"404", http.StatusNotFound, `{}`, "nothing matched"},
		{"409 non dup", http.StatusConflict, `{"message":"other"}`, "conflicts"},
		{"422", http.StatusUnprocessableEntity, `{"message":"bad"}`, "couldn't be saved"},
		{"500 generic", http.StatusInternalServerError, `{"message":"oops"}`, "something went wrong"},
		{"418 default", 418, `{}`, "didn't succeed"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := apiError("GET", "/rest/v1/x", "status", tc.status, []byte(tc.raw))
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.substr)
		})
	}
}

func TestAPIError_42501(t *testing.T) {
	t.Setenv(config.EnvDebug, "")
	err := apiError("PATCH", "/x", "500", http.StatusInternalServerError, []byte(`{"code":"42501"}`))
	require.Contains(t, err.Error(), "permission denied")
}
