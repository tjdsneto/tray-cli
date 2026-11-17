package api

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/tjdsneto/tray-cli/internal/bar"
)

type Handler struct {
	token    string
	src      Source
	interval time.Duration
	mu       sync.Mutex
	snap     bar.Snapshot
}

func NewHandler(token string, src Source, interval time.Duration) *Handler {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	return &Handler{token: token, src: src, interval: interval}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(r) {
		WriteError(w, http.StatusUnauthorized, "unavailable", "unauthorized")
		return
	}
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/v1/session":
		h.session(w, r)
	case r.Method == http.MethodGet && r.URL.Path == "/v1/snapshot":
		h.snapshot(w, r)
	default:
		WriteError(w, http.StatusNotFound, "not_found", "not found")
	}
}

func (h *Handler) authorized(r *http.Request) bool {
	const prefix = "Bearer "
	got := r.Header.Get("Authorization")
	if !strings.HasPrefix(got, prefix) {
		return false
	}
	token := got[len(prefix):]
	return subtle.ConstantTimeCompare([]byte(token), []byte(h.token)) == 1
}

func (h *Handler) session(w http.ResponseWriter, r *http.Request) {
	ok, err := h.src.SignedIn(r.Context())
	if err != nil {
		WriteError(w, http.StatusBadGateway, "unavailable", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, Session{SignedIn: ok})
}

func (h *Handler) snapshot(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	snap := SnapshotFrom(h.snap)
	h.mu.Unlock()
	writeJSON(w, http.StatusOK, snap)
}

func (h *Handler) Rebuild(ctx Ctx) {
	local, remote, remoteErr := h.src.Attention(ctx)
	h.mu.Lock()
	defer h.mu.Unlock()
	if remoteErr != nil {
		remote = bar.RemoteItemsFromSnapshot(h.snap)
		h.snap = bar.BuildSnapshot(append(append([]bar.Item{}, local...), remote...), "remote unavailable")
		return
	}
	h.snap = bar.BuildSnapshot(append(append([]bar.Item{}, local...), remote...), "")
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
