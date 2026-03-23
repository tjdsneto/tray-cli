package postgrest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tjdsneto/tray-cli/internal/adapters/postgrest/pghttp"
	"github.com/tjdsneto/tray-cli/internal/domain"
	supabasehttp "github.com/tjdsneto/tray-cli/internal/supabase"
)

func TestItemsListPath_filtersAndOrder(t *testing.T) {
	t.Parallel()
	p := itemsListPath(domain.ListItemsQuery{
		ItemID:       "i1",
		TrayID:       "t1",
		Status:       "pending",
		OrderCreated: "asc",
	})
	require.Contains(t, p, "id=eq.i1")
	require.Contains(t, p, "tray_id=eq.t1")
	require.Contains(t, p, "status=eq.pending")
	require.Contains(t, p, "order=created_at.asc")
}

func TestItemsListPath_defaultOrderDesc(t *testing.T) {
	t.Parallel()
	p := itemsListPath(domain.ListItemsQuery{})
	require.Contains(t, p, "order=created_at.desc")
}

func TestItemsCreatePath_select(t *testing.T) {
	t.Parallel()
	p := itemsCreatePath()
	q, err := url.ParseQuery(strings.TrimPrefix(p, "/rest/v1/items?"))
	require.NoError(t, err)
	require.Equal(t, itemSelectColumns, q.Get("select"))
}

func TestItemsOutboxPath(t *testing.T) {
	t.Parallel()
	p := itemsOutboxPath("user-1")
	require.Contains(t, p, "source_user_id=eq.user-1")
	require.Contains(t, p, "trays%28owner_id%29") // trays(owner_id) encoded
}

func TestItemsPatchPath(t *testing.T) {
	t.Parallel()
	p := itemsPatchPath("  abc  ")
	require.Contains(t, p, "id=eq.abc")
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
