package api_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tjdsneto/tray-cli/internal/api"
	"github.com/tjdsneto/tray-cli/internal/bar"
)

type fakeSource struct {
	signedIn  bool
	local     []bar.Item
	remote    []bar.Item
	remoteErr error
	trays     []api.Tray
	items     []api.Item
	actErr    error
}

func (f *fakeSource) SignedIn(api.Ctx) (bool, error) { return f.signedIn, nil }
func (f *fakeSource) LocalAttention(api.Ctx) ([]bar.Item, error) {
	return f.local, nil
}
func (f *fakeSource) Attention(api.Ctx) ([]bar.Item, []bar.Item, error) {
	return f.local, f.remote, f.remoteErr
}
func (f *fakeSource) Trays(api.Ctx) ([]api.Tray, error)         { return f.trays, nil }
func (f *fakeSource) Items(api.Ctx, string) ([]api.Item, error) { return f.items, nil }
func (f *fakeSource) Item(ctx api.Ctx, id string) (api.Item, error) {
	for _, it := range f.items {
		if it.ID == id {
			return it, nil
		}
	}
	return api.Item{}, api.ErrNotFound
}
func (f *fakeSource) Add(api.Ctx, api.AddRequest) (api.Item, error) {
	return api.Item{}, nil
}
func (f *fakeSource) Act(api.Ctx, string, string, api.ActBody) error { return f.actErr }

func TestSnapshot_requiresBearerAndKeepsLastRemote(t *testing.T) {
	t.Parallel()
	src := &fakeSource{signedIn: true, local: []bar.Item{{
		Scope: bar.ScopeLocal, TrayRef: "notes", TrayName: "notes", ItemID: "l1", Title: "L", CreatedAt: time.Unix(1, 0).UTC(),
	}}, remote: []bar.Item{{
		Scope: bar.ScopeRemote, TrayRef: "inbox", TrayName: "inbox", ItemID: "r1", Title: "R", CreatedAt: time.Unix(2, 0).UTC(),
	}}}
	h := api.NewHandler("secret", src, time.Hour)
	h.Rebuild(t.Context())

	denied := httptest.NewRecorder()
	h.ServeHTTP(denied, httptest.NewRequest(http.MethodGet, "/v1/snapshot", nil))
	require.Equal(t, http.StatusUnauthorized, denied.Code)

	src.remoteErr = api.ErrRemote
	h.Rebuild(t.Context())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/snapshot", nil)
	req.Header.Set("Authorization", "Bearer secret")
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	var snap api.Snapshot
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &snap))
	require.Equal(t, "remote unavailable", snap.RemoteStatus)
	require.Equal(t, 2, snap.Badge)
}

func TestItems_listOmitsBodyAndGetIncludesIt(t *testing.T) {
	t.Parallel()
	body := "full notes"
	src := &fakeSource{items: []api.Item{
		{ID: "c", Scope: "remote", TrayRef: "inbox", Title: "New", Status: "pending", CreatedAt: time.Unix(3, 0).UTC(), Body: &body, Actions: []string{"accept", "decline", "snooze"}},
		{ID: "b", Scope: "remote", TrayRef: "inbox", Title: "Old", Status: "completed", CreatedAt: time.Unix(1, 0).UTC()},
	}}
	h := api.NewHandler("secret", src, time.Hour)
	rec := getAuth(h, "/v1/items?tray=inbox&filter=open")
	require.Equal(t, http.StatusOK, rec.Code)
	var list []api.Item
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &list))
	require.Equal(t, []string{"c"}, ids(list))
	require.Nil(t, list[0].Body)

	one := getAuth(h, "/v1/items/c")
	require.Equal(t, http.StatusOK, one.Code)
	var item api.Item
	require.NoError(t, json.Unmarshal(one.Body.Bytes(), &item))
	require.Equal(t, "full notes", *item.Body)

	missing := getAuth(h, "/v1/items/nope")
	require.Equal(t, http.StatusNotFound, missing.Code)
}

func TestRebuildLocal_publishesDiskItems(t *testing.T) {
	t.Parallel()
	src := &fakeSource{local: []bar.Item{{
		Scope: bar.ScopeLocal, TrayRef: "notes", TrayName: "notes", ItemID: "l1", Title: "L", CreatedAt: time.Unix(1, 0).UTC(),
	}}}
	h := api.NewHandler("secret", src, time.Hour)
	h.RebuildLocal(t.Context())

	snap := getAuth(h, "/v1/snapshot")
	var body api.Snapshot
	require.NoError(t, json.Unmarshal(snap.Body.Bytes(), &body))
	require.Equal(t, 1, body.Badge)
	require.Equal(t, "l1", body.Sections[0].Rows[0].ItemID)
	require.Empty(t, body.RemoteStatus)
}

func TestRebuild_localStoreErrorKeepsSnapshot(t *testing.T) {
	t.Parallel()
	src := &fakeSource{local: []bar.Item{{
		Scope: bar.ScopeLocal, TrayRef: "notes", TrayName: "notes", ItemID: "l1", Title: "L", CreatedAt: time.Unix(1, 0).UTC(),
	}}}
	h := api.NewHandler("secret", src, time.Hour)
	h.Rebuild(t.Context())

	src.local = nil
	src.remote = nil
	src.remoteErr = errors.New("disk")
	h.Rebuild(t.Context())

	snap := getAuth(h, "/v1/snapshot")
	var body api.Snapshot
	require.NoError(t, json.Unmarshal(snap.Body.Bytes(), &body))
	require.Equal(t, 1, body.Badge)
	require.Equal(t, "l1", body.Sections[0].Rows[0].ItemID)
}

func TestAct_failureDoesNotRebuildSnapshot(t *testing.T) {
	t.Parallel()
	src := &fakeSource{local: []bar.Item{{
		Scope: bar.ScopeLocal, TrayRef: "notes", TrayName: "notes", ItemID: "l1", Title: "L", CreatedAt: time.Unix(1, 0).UTC(),
	}}}
	src.actErr = errors.New("nope")
	h := api.NewHandler("secret", src, time.Hour)
	h.Rebuild(t.Context())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/items/l1/complete", nil)
	req.Header.Set("Authorization", "Bearer secret")
	h.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadGateway, rec.Code)

	snap := getAuth(h, "/v1/snapshot")
	var body api.Snapshot
	require.NoError(t, json.Unmarshal(snap.Body.Bytes(), &body))
	require.Equal(t, 1, body.Badge)
	require.Equal(t, "l1", body.Sections[0].Rows[0].ItemID)
}

func TestSnapshot_optionsPreflightSkipsBearer(t *testing.T) {
	t.Parallel()
	h := api.NewHandler("secret", &fakeSource{}, time.Hour)

	preflight := httptest.NewRecorder()
	h.ServeHTTP(preflight, httptest.NewRequest(http.MethodOptions, "/v1/snapshot", nil))
	require.Equal(t, http.StatusNoContent, preflight.Code)
	require.Empty(t, preflight.Body.String())
	require.Equal(t, "*", preflight.Header().Get("Access-Control-Allow-Origin"))
	require.Equal(t, "Authorization, Content-Type", preflight.Header().Get("Access-Control-Allow-Headers"))
	require.Equal(t, "GET, POST, OPTIONS", preflight.Header().Get("Access-Control-Allow-Methods"))

	rec := getAuth(h, "/v1/snapshot")
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "*", rec.Header().Get("Access-Control-Allow-Origin"))
}

func getAuth(h http.Handler, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("Authorization", "Bearer secret")
	h.ServeHTTP(rec, req)
	return rec
}
