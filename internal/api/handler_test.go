package api_test

import (
	"encoding/json"
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
}

func (f *fakeSource) SignedIn(api.Ctx) (bool, error) { return f.signedIn, nil }
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
func (f *fakeSource) Act(api.Ctx, string, string, api.ActBody) error { return nil }

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

func getAuth(h http.Handler, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("Authorization", "Bearer secret")
	h.ServeHTTP(rec, req)
	return rec
}
