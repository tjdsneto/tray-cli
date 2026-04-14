package postgrest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

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

func TestItemsListPath_trayIDIn(t *testing.T) {
	t.Parallel()
	p := itemsListPath(domain.ListItemsQuery{
		TrayIDIn: []string{"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"},
	})
	require.Contains(t, p, "tray_id=in.")
	require.Contains(t, p, "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	require.Contains(t, p, "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
}

func TestItemsListPath_updatedAfter(t *testing.T) {
	t.Parallel()
	ts := time.Date(2026, 3, 24, 13, 30, 45, 123000000, time.UTC)
	p := itemsListPath(domain.ListItemsQuery{UpdatedAfter: &ts})
	q, err := url.ParseQuery(strings.TrimPrefix(p, "/rest/v1/items?"))
	require.NoError(t, err)
	require.Equal(t, "gt."+ts.UTC().Format(time.RFC3339Nano), q.Get("updated_at"))
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

func TestItemService_Add_ownerGetsAccepted(t *testing.T) {
	var postBody addItemRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/rest/v1/trays" && r.Method == http.MethodGet:
			_ = json.NewEncoder(w).Encode([]map[string]string{{"owner_id": "u-owner"}})
		case r.URL.Path == "/rest/v1/items" && r.Method == http.MethodPost:
			require.NoError(t, json.NewDecoder(r.Body).Decode(&postBody))
			require.Equal(t, "accepted", postBody.Status)
			_, _ = w.Write([]byte(`[{"id":"i1","tray_id":"t1","source_user_id":"u-owner","title":"x","status":"accepted","created_at":"2026-03-20T12:00:00Z","updated_at":"2026-03-20T12:00:00Z"}]`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)
	c, err := supabasehttp.NewClient(srv.URL, "anon", srv.Client())
	require.NoError(t, err)
	svc := newItemService(pghttp.New(c))
	it, err := svc.Add(context.Background(), domain.Session{AccessToken: "tok", UserID: "u-owner"}, "t1", "x", nil)
	require.NoError(t, err)
	require.Equal(t, "accepted", it.Status)
}

func TestItemService_Add_contributorGetsPending(t *testing.T) {
	var postBody addItemRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/rest/v1/trays" && r.Method == http.MethodGet:
			_ = json.NewEncoder(w).Encode([]map[string]string{{"owner_id": "owner-1"}})
		case r.URL.Path == "/rest/v1/items" && r.Method == http.MethodPost:
			require.NoError(t, json.NewDecoder(r.Body).Decode(&postBody))
			require.Equal(t, "pending", postBody.Status)
			_, _ = w.Write([]byte(`[{"id":"i2","tray_id":"t1","source_user_id":"member","title":"y","status":"pending","created_at":"2026-03-20T12:00:00Z","updated_at":"2026-03-20T12:00:00Z"}]`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)
	c, err := supabasehttp.NewClient(srv.URL, "anon", srv.Client())
	require.NoError(t, err)
	svc := newItemService(pghttp.New(c))
	it, err := svc.Add(context.Background(), domain.Session{AccessToken: "tok", UserID: "member"}, "t1", "y", nil)
	require.NoError(t, err)
	require.Equal(t, "pending", it.Status)
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
