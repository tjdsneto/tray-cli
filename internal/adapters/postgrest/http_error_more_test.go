package postgrest

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tjdsneto/tray-cli/internal/config"
)

func TestHTTPAPIError_statuses(t *testing.T) {
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
			err := httpAPIError("GET", "/rest/v1/x", "status", tc.status, []byte(tc.raw))
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.substr)
		})
	}
}

func TestHTTPAPIError_42501(t *testing.T) {
	t.Setenv(config.EnvDebug, "")
	err := httpAPIError("PATCH", "/x", "500", http.StatusInternalServerError, []byte(`{"code":"42501"}`))
	require.Contains(t, err.Error(), "permission denied")
}
