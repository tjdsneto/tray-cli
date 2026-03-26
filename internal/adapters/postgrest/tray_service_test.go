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

func TestTraysCreatePath(t *testing.T) {
	t.Parallel()
	require.Equal(t, "/rest/v1/trays", traysCreatePath())
}

func TestTraysListMinePath(t *testing.T) {
	t.Parallel()
	p := traysListMinePath()
	require.Contains(t, p, "order=name.asc")
	require.Contains(t, p, "items%28count%29")
	q, err := url.ParseQuery(strings.TrimPrefix(p, "/rest/v1/trays?"))
	require.NoError(t, err)
	require.Contains(t, q.Get("select"), "owner_id")
}

func TestTraysListOwnedPath(t *testing.T) {
	t.Parallel()
	p := traysListOwnedPath("user-uuid")
	require.Contains(t, p, "owner_id=eq.user-uuid")
	require.Contains(t, p, "order=name.asc")
}

func TestTrayMembersListForUserPath(t *testing.T) {
	t.Parallel()
	p := trayMembersListForUserPath("member-uid")
	require.Contains(t, p, "user_id=eq.member-uid")
	require.Contains(t, p, "trays%28id")
}

func TestJoinTrayRPCPath(t *testing.T) {
	t.Parallel()
	require.Equal(t, "/rest/v1/rpc/join_tray", joinTrayRPCPath())
}

func TestTraysByIDPath(t *testing.T) {
	t.Parallel()
	p := traysByIDPath("  uuid-1  ")
	require.Contains(t, p, "id=eq.uuid-1")
}

func TestTrayMembersListPath(t *testing.T) {
	t.Parallel()
	p := trayMembersListPath("t99")
	require.Contains(t, p, "tray_id=eq.t99")
	require.Contains(t, p, "order=joined_at.asc")
	q, err := url.ParseQuery(strings.TrimPrefix(p, "/rest/v1/tray_members?"))
	require.NoError(t, err)
	require.Equal(t, trayMemberSelectColumns, q.Get("select"))
}

func TestTrayMembersDeletePath(t *testing.T) {
	t.Parallel()
	p := trayMembersDeletePath("trayA", "userB")
	require.Contains(t, p, "tray_id=eq.trayA")
	require.Contains(t, p, "user_id=eq.userB")
}
