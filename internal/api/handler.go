package api

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
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
	case r.Method == http.MethodGet && r.URL.Path == "/v1/trays":
		h.trays(w, r)
	case r.Method == http.MethodGet && r.URL.Path == "/v1/items":
		h.items(w, r)
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/v1/items/"):
		h.item(w, r)
	case r.Method == http.MethodPost && r.URL.Path == "/v1/items":
		h.add(w, r)
	case r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/v1/items/"):
		h.act(w, r)
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

func (h *Handler) trays(w http.ResponseWriter, r *http.Request) {
	trays, err := h.src.Trays(r.Context())
	if err != nil {
		writeSourceErr(w, err)
		return
	}
	if trays == nil {
		trays = []Tray{}
	}
	writeJSON(w, http.StatusOK, trays)
}

func (h *Handler) items(w http.ResponseWriter, r *http.Request) {
	tray := r.URL.Query().Get("tray")
	filter := r.URL.Query().Get("filter")
	if filter == "" {
		filter = "open"
	}
	items, err := h.src.Items(r.Context(), tray)
	if err != nil {
		writeSourceErr(w, err)
		return
	}
	out := SelectItems(items, filter)
	for i := range out {
		out[i].Body = nil
		if out[i].Actions == nil {
			out[i].Actions = ActionsFor(out[i].Scope, out[i].Status)
		}
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) add(w http.ResponseWriter, r *http.Request) {
	var req AddRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid", "invalid json")
		return
	}
	if err := ValidateAdd(req); err != nil {
		writeSourceErr(w, err)
		return
	}
	item, err := h.src.Add(r.Context(), req)
	if err != nil {
		writeSourceErr(w, err)
		return
	}
	h.Rebuild(r.Context())
	item.Actions = ActionsFor(item.Scope, item.Status)
	writeJSON(w, http.StatusCreated, item)
}

func (h *Handler) act(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(r.URL.Path, "/v1/items/")
	id, action, ok := strings.Cut(rest, "/")
	if !ok || id == "" || action == "" {
		WriteError(w, http.StatusNotFound, "not_found", "not found")
		return
	}
	var body ActBody
	if r.ContentLength != 0 {
		var raw struct {
			Reason string `json:"reason"`
			Until  string `json:"until"`
		}
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			WriteError(w, http.StatusBadRequest, "invalid", "invalid json")
			return
		}
		body.Reason, body.Until = raw.Reason, raw.Until
	}
	if err := h.src.Act(r.Context(), id, action, body); err != nil {
		writeSourceErr(w, err)
		return
	}
	h.Rebuild(r.Context())
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) item(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/v1/items/")
	if id == "" || strings.Contains(id, "/") {
		WriteError(w, http.StatusNotFound, "not_found", "not found")
		return
	}
	item, err := h.src.Item(r.Context(), id)
	if err != nil {
		writeSourceErr(w, err)
		return
	}
	if item.Actions == nil {
		item.Actions = ActionsFor(item.Scope, item.Status)
	}
	writeJSON(w, http.StatusOK, item)
}

func writeSourceErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		WriteError(w, http.StatusNotFound, "not_found", "not found")
	case errors.Is(err, ErrSignedOut):
		WriteError(w, http.StatusUnauthorized, "signed_out", "signed out")
	case errors.Is(err, ErrRemote):
		WriteError(w, http.StatusBadGateway, "remote_unavailable", "remote unavailable")
	default:
		var inv invalidError
		if errors.As(err, &inv) {
			WriteError(w, http.StatusBadRequest, "invalid", inv.Error())
			return
		}
		WriteError(w, http.StatusBadGateway, "unavailable", err.Error())
	}
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
