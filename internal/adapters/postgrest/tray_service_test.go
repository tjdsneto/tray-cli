package postgrest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tjdsneto/tray-cli/internal/domain"
	supabasehttp "github.com/tjdsneto/tray-cli/internal/supabase"
)

func TestTrayService_ListMine(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/rest/v1/trays", r.URL.Path)
		require.Equal(t, http.MethodGet, r.Method)
		require.Contains(t, r.URL.RawQuery, "items")
		_ = json.NewEncoder(w).Encode([]trayRow{
			{
				ID: "a", OwnerID: "u1", Name: "work", CreatedAt: "2026-03-20T12:00:00Z",
				Items: []trayItemsCount{{Count: 3}},
			},
		})
	}))
	t.Cleanup(srv.Close)

	c, err := supabasehttp.NewClient(srv.URL, "anon", srv.Client())
	require.NoError(t, err)
	svc := newTrayService(newClient(c))

	trays, err := svc.ListMine(context.Background(), domain.Session{AccessToken: "jwt"})
	require.NoError(t, err)
	require.Len(t, trays, 1)
	require.Equal(t, "work", trays[0].Name)
	require.Equal(t, 3, trays[0].ItemCount)
}

func TestTrayService_Join(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/rest/v1/rpc/join_tray", r.URL.Path)
		_, _ = w.Write([]byte(`"9b1d4c8e-7a2f-4e3d-9c1b-000000000001"`))
	}))
	t.Cleanup(srv.Close)

	c, err := supabasehttp.NewClient(srv.URL, "anon", srv.Client())
	require.NoError(t, err)
	svc := newTrayService(newClient(c))

	id, err := svc.Join(context.Background(), domain.Session{AccessToken: "jwt"}, "tok")
	require.NoError(t, err)
	require.Equal(t, "9b1d4c8e-7a2f-4e3d-9c1b-000000000001", id)
}

func TestTrayService_Create_requiresUserID(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(srv.Close)
	c, err := supabasehttp.NewClient(srv.URL, "anon", srv.Client())
	require.NoError(t, err)
	svc := newTrayService(newClient(c))
	_, err = svc.Create(context.Background(), domain.Session{AccessToken: "j"}, "x", nil)
	require.Error(t, err)
}

func TestItemService_stubs(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(srv.Close)
	c, err := supabasehttp.NewClient(srv.URL, "anon", srv.Client())
	require.NoError(t, err)
	svc := newItemService(newClient(c))
	ctx := context.Background()
	s := domain.Session{AccessToken: "x", UserID: "u"}

	_, err = svc.Add(ctx, s, "t", "title", nil)
	require.ErrorIs(t, err, domain.ErrNotImplemented)
}

func TestTrayService_UpdateName_notImplemented(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(srv.Close)
	c, err := supabasehttp.NewClient(srv.URL, "anon", srv.Client())
	require.NoError(t, err)
	svc := newTrayService(newClient(c))
	err = svc.UpdateName(context.Background(), domain.Session{}, "id", "n")
	require.ErrorIs(t, err, domain.ErrNotImplemented)
}
