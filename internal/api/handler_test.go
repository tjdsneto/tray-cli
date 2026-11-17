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
}

func (f *fakeSource) SignedIn(api.Ctx) (bool, error) { return f.signedIn, nil }
func (f *fakeSource) Attention(api.Ctx) ([]bar.Item, []bar.Item, error) {
	return f.local, f.remote, f.remoteErr
}
func (f *fakeSource) Trays(api.Ctx) ([]api.Tray, error)         { return nil, nil }
func (f *fakeSource) Items(api.Ctx, string) ([]api.Item, error) { return nil, nil }
func (f *fakeSource) Item(api.Ctx, string) (api.Item, error) {
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
