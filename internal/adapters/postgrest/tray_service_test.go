package postgrest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tjdsneto/tray-cli/internal/adapters/postgrest/pghttp"
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
	svc := newTrayService(pghttp.New(c))

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
	svc := newTrayService(pghttp.New(c))

	id, err := svc.Join(context.Background(), domain.Session{AccessToken: "jwt"}, "tok")
	require.NoError(t, err)
	require.Equal(t, "9b1d4c8e-7a2f-4e3d-9c1b-000000000001", id)
}

func TestTrayService_Create_requiresUserID(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(srv.Close)
	c, err := supabasehttp.NewClient(srv.URL, "anon", srv.Client())
	require.NoError(t, err)
	svc := newTrayService(pghttp.New(c))
	_, err = svc.Create(context.Background(), domain.Session{AccessToken: "j"}, "x", nil)
	require.Error(t, err)
}

func TestItemService_List_withItemID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/rest/v1/items", r.URL.Path)
		require.Equal(t, http.MethodGet, r.Method)
		require.Contains(t, r.URL.RawQuery, "id=eq.item-uuid")
		_ = json.NewEncoder(w).Encode([]any{})
	}))
	t.Cleanup(srv.Close)
	c, err := supabasehttp.NewClient(srv.URL, "anon", srv.Client())
	require.NoError(t, err)
	svc := newItemService(pghttp.New(c))
	ctx := context.Background()
	s := domain.Session{AccessToken: "x", UserID: "u"}

	items, err := svc.List(ctx, s, domain.ListItemsQuery{ItemID: "item-uuid"})
	require.NoError(t, err)
	require.Empty(t, items)
}

func TestItemService_Add_notFound(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(srv.Close)
	c, err := supabasehttp.NewClient(srv.URL, "anon", srv.Client())
	require.NoError(t, err)
	svc := newItemService(pghttp.New(c))
	ctx := context.Background()
	s := domain.Session{AccessToken: "x", UserID: "u"}

	_, err = svc.Add(ctx, s, "00000000-0000-0000-0000-000000000001", "title", nil)
	require.Error(t, err)
}

func TestItemService_Update_emptyPatch(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(srv.Close)
	c, err := supabasehttp.NewClient(srv.URL, "anon", srv.Client())
	require.NoError(t, err)
	svc := newItemService(pghttp.New(c))
	ctx := context.Background()
	s := domain.Session{AccessToken: "x", UserID: "u"}

	err = svc.Update(ctx, s, "00000000-0000-0000-0000-000000000001", domain.ItemPatch{})
	require.Error(t, err)
}

func TestItemService_Update_patch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPatch, r.Method)
		require.Contains(t, r.URL.RawQuery, "id=eq.")
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)
	c, err := supabasehttp.NewClient(srv.URL, "anon", srv.Client())
	require.NoError(t, err)
	svc := newItemService(pghttp.New(c))
	ctx := context.Background()
	s := domain.Session{AccessToken: "x", UserID: "u"}
	st := "accepted"
	err = svc.Update(ctx, s, "00000000-0000-0000-0000-000000000002", domain.ItemPatch{Status: &st})
	require.NoError(t, err)
}

func TestItemService_Delete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodDelete, r.Method)
		require.Contains(t, r.URL.RawQuery, "id=eq.")
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)
	c, err := supabasehttp.NewClient(srv.URL, "anon", srv.Client())
	require.NoError(t, err)
	svc := newItemService(pghttp.New(c))
	err = svc.Delete(context.Background(), domain.Session{AccessToken: "x", UserID: "u"}, "00000000-0000-0000-0000-000000000099")
	require.NoError(t, err)
}

func TestTrayService_UpdateName(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPatch, r.Method)
		require.Contains(t, r.URL.RawQuery, "id=eq.")
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)
	c, err := supabasehttp.NewClient(srv.URL, "anon", srv.Client())
	require.NoError(t, err)
	svc := newTrayService(pghttp.New(c))
	err = svc.UpdateName(context.Background(), domain.Session{AccessToken: "x", UserID: "u"}, "00000000-0000-0000-0000-000000000001", "new")
	require.NoError(t, err)
}

func TestTrayService_ListMembers(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/rest/v1/tray_members", r.URL.Path)
		require.Contains(t, r.URL.RawQuery, "tray_id=eq.")
		_ = json.NewEncoder(w).Encode([]trayMemberRow{{
			ID: "m1", TrayID: "t1", UserID: "u2",
			JoinedAt: "2026-03-20T12:00:00Z",
		}})
	}))
	t.Cleanup(srv.Close)
	c, err := supabasehttp.NewClient(srv.URL, "anon", srv.Client())
	require.NoError(t, err)
	svc := newTrayService(pghttp.New(c))
	members, err := svc.ListMembers(context.Background(), domain.Session{AccessToken: "x"}, "t1")
	require.NoError(t, err)
	require.Len(t, members, 1)
	require.Equal(t, "u2", members[0].UserID)
}
