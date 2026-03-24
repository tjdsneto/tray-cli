package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tjdsneto/tray-cli/internal/credentials"
)

func TestRefreshTokens(t *testing.T) {
	var gotBody map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/auth/v1/token", r.URL.Path)
		require.Equal(t, "refresh_token", r.URL.Query().Get("grant_type"))
		require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "new-access",
			"refresh_token": "new-refresh",
			"user":          map[string]string{"id": "user-1"},
		})
	}))
	defer srv.Close()

	access, refresh, uid, err := RefreshTokens(context.Background(), srv.URL, "anon", "old-refresh", srv.Client())
	require.NoError(t, err)
	require.Equal(t, "new-access", access)
	require.Equal(t, "new-refresh", refresh)
	require.Equal(t, "user-1", uid)
	require.Equal(t, "old-refresh", gotBody["refresh_token"])
}

func TestEnsureFreshCredentials_skipsWithoutRefreshToken(t *testing.T) {
	dir := t.TempDir()
	f := credentials.File{AccessToken: "a", UserID: "u"}
	out, err := EnsureFreshCredentials(context.Background(), "http://x", "k", nil, dir, f)
	require.NoError(t, err)
	require.Equal(t, f, out)
}

func TestEnsureFreshCredentials_refreshesWhenExpiring(t *testing.T) {
	past := float64(time.Now().Add(-2 * time.Hour).Unix())
	accessTok := "eyJhbGciOiJub25lIn0." + mustB64JSON(t, map[string]any{"exp": past}) + ".x"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "refreshed",
			"refresh_token": "rotated",
			"user":          map[string]string{"id": "uid-2"},
		})
	}))
	defer srv.Close()

	dir := t.TempDir()
	f := credentials.File{
		AccessToken:  accessTok,
		RefreshToken: "rt",
		UserID:       "old",
	}
	out, err := EnsureFreshCredentials(context.Background(), srv.URL, "anon", srv.Client(), dir, f)
	require.NoError(t, err)
	require.Equal(t, "refreshed", out.AccessToken)
	require.Equal(t, "rotated", out.RefreshToken)
	require.Equal(t, "uid-2", out.UserID)

	loaded, err := credentials.Load(dir)
	require.NoError(t, err)
	require.Equal(t, out.AccessToken, loaded.AccessToken)
}

func TestEnsureFreshCredentials_noRefreshWhenValid(t *testing.T) {
	far := float64(time.Now().Add(1 * time.Hour).Unix())
	accessTok := "eyJhbGciOiJub25lIn0." + mustB64JSON(t, map[string]any{"exp": far}) + ".x"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("server should not be called")
	}))
	defer srv.Close()

	dir := t.TempDir()
	f := credentials.File{AccessToken: accessTok, RefreshToken: "rt", UserID: "u"}
	out, err := EnsureFreshCredentials(context.Background(), srv.URL, "anon", srv.Client(), dir, f)
	require.NoError(t, err)
	require.Equal(t, f.AccessToken, out.AccessToken)
}
