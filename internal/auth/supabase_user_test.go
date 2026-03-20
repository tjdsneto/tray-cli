package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFetchUser(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/auth/v1/user", r.URL.Path)
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "anon", r.Header.Get("apikey"))
		require.Equal(t, "Bearer jwt", r.Header.Get("Authorization"))
		_, _ = w.Write([]byte(`{"id":"11111111-1111-1111-1111-111111111111","email":"a@b.c"}`))
	}))
	t.Cleanup(srv.Close)

	u, err := FetchUser(context.Background(), srv.URL, "anon", "jwt", srv.Client())
	require.NoError(t, err)
	require.Equal(t, "11111111-1111-1111-1111-111111111111", u.ID)
	require.Equal(t, "a@b.c", u.Email)
}

func TestFetchUser_errors(t *testing.T) {
	_, err := FetchUser(context.Background(), "", "k", "t", nil)
	require.Error(t, err)
	_, err = FetchUser(context.Background(), "http://x", "", "t", nil)
	require.Error(t, err)
}
