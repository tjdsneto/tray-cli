package supabase

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewClient_errors(t *testing.T) {
	_, err := NewClient("", "key", nil)
	require.Error(t, err)
	_, err = NewClient("https://x.supabase.co", "", nil)
	require.Error(t, err)
}

func TestClient_NewRequest_headers(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "GET", r.Method)
		require.Equal(t, "/rest/v1/trays", r.URL.Path)
		require.Equal(t, "anon-secret", r.Header.Get("apikey"))
		require.Equal(t, "Bearer user-jwt", r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	c, err := NewClient(srv.URL, "anon-secret", srv.Client())
	require.NoError(t, err)

	req, err := c.NewRequest(context.Background(), http.MethodGet, "/rest/v1/trays", nil, "user-jwt")
	require.NoError(t, err)

	resp, err := c.HTTPClient.Do(req)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.Equal(t, http.StatusOK, resp.StatusCode)
}
