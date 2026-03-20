package backend

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tjdsneto/tray-cli/internal/supabase"
)

func TestSupabaseBackend_ListMyTrays(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/rest/v1/trays", r.URL.Path)
		require.Equal(t, http.MethodGet, r.Method)
		_ = json.NewEncoder(w).Encode([]trayJSON{
			{ID: "a", OwnerID: "u1", Name: "work", CreatedAt: "2026-03-20T12:00:00Z"},
		})
	}))
	t.Cleanup(srv.Close)

	c, err := supabase.NewClient(srv.URL, "anon", srv.Client())
	require.NoError(t, err)
	b := NewSupabase(c)

	trays, err := b.ListMyTrays(context.Background(), Session{AccessToken: "jwt"})
	require.NoError(t, err)
	require.Len(t, trays, 1)
	require.Equal(t, "work", trays[0].Name)
}

func TestSupabaseBackend_JoinTray(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/rest/v1/rpc/join_tray", r.URL.Path)
		_, _ = w.Write([]byte(`"9b1d4c8e-7a2f-4e3d-9c1b-000000000001"`))
	}))
	t.Cleanup(srv.Close)

	c, err := supabase.NewClient(srv.URL, "anon", srv.Client())
	require.NoError(t, err)
	b := NewSupabase(c)

	id, err := b.JoinTray(context.Background(), Session{AccessToken: "jwt"}, "tok")
	require.NoError(t, err)
	require.Equal(t, "9b1d4c8e-7a2f-4e3d-9c1b-000000000001", id)
}

func TestParseCreatedTray_arrayOrObject(t *testing.T) {
	row, err := parseCreatedTray([]byte(`[{"id":"x","owner_id":"o","name":"n","created_at":"2026-01-02T15:04:05Z"}]`))
	require.NoError(t, err)
	require.Equal(t, "x", row.ID)

	row, err = parseCreatedTray([]byte(`{"id":"y","owner_id":"o","name":"n2","created_at":"2026-01-02T15:04:05Z"}`))
	require.NoError(t, err)
	require.Equal(t, "y", row.ID)
}

func TestSupabaseBackend_CreateTray_requiresUserID(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(srv.Close)
	c, err := supabase.NewClient(srv.URL, "anon", srv.Client())
	require.NoError(t, err)
	b := NewSupabase(c)
	_, err = b.CreateTray(context.Background(), Session{AccessToken: "j"}, "x", nil)
	require.Error(t, err)
}

func TestStubMethods_NotImplemented(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(srv.Close)
	c, err := supabase.NewClient(srv.URL, "anon", srv.Client())
	require.NoError(t, err)
	b := NewSupabase(c)
	ctx := context.Background()
	s := Session{AccessToken: "x", UserID: "u"}

	_, err = b.AddItem(ctx, s, "t", "title", nil)
	require.ErrorIs(t, err, ErrNotImplemented)

	err = b.UpdateTrayName(ctx, s, "id", "n")
	require.ErrorIs(t, err, ErrNotImplemented)
}
