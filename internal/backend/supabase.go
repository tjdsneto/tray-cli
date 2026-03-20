package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/tjdsneto/tray-cli/internal/supabase"
)

// supabaseBackend implements Backend using PostgREST (Supabase Data API).
type supabaseBackend struct {
	client *supabase.Client
}

// NewSupabase returns a Backend backed by Supabase PostgREST.
func NewSupabase(c *supabase.Client) Backend {
	return &supabaseBackend{client: c}
}

// DialSupabase builds a Client and Backend in one step (typical CLI wiring).
func DialSupabase(rawURL, anonKey string, httpClient *http.Client) (Backend, error) {
	c, err := supabase.NewClient(rawURL, anonKey, httpClient)
	if err != nil {
		return nil, err
	}
	return NewSupabase(c), nil
}

var _ Backend = (*supabaseBackend)(nil)

// --- Trays (implemented) ---

func (b *supabaseBackend) CreateTray(ctx context.Context, sess Session, name string, inviteToken *string) (*Tray, error) {
	if strings.TrimSpace(sess.UserID) == "" {
		return nil, fmt.Errorf("backend: session missing UserID (set after login)")
	}
	body := map[string]any{
		"name":     name,
		"owner_id": sess.UserID,
	}
	if inviteToken != nil {
		body["invite_token"] = *inviteToken
	}
	raw, err := b.request(ctx, sess, http.MethodPost, "/rest/v1/trays", body, hdrPreferRepresentation())
	if err != nil {
		return nil, err
	}
	row, err := parseCreatedTray(raw)
	if err != nil {
		return nil, err
	}
	return trayFromJSON(*row)
}

func (b *supabaseBackend) ListMyTrays(ctx context.Context, sess Session) ([]Tray, error) {
	path := "/rest/v1/trays?select=id,owner_id,name,invite_token,created_at&order=name.asc"
	var rows []trayJSON
	if err := b.doJSON(ctx, sess, http.MethodGet, path, nil, &rows, nil); err != nil {
		return nil, err
	}
	out := make([]Tray, 0, len(rows))
	for _, r := range rows {
		t, err := trayFromJSON(r)
		if err != nil {
			return nil, err
		}
		out = append(out, *t)
	}
	return out, nil
}

func (b *supabaseBackend) JoinTray(ctx context.Context, sess Session, inviteToken string) (string, error) {
	body := map[string]string{"p_invite_token": strings.TrimSpace(inviteToken)}
	var trayID string
	if err := b.doJSON(ctx, sess, http.MethodPost, "/rest/v1/rpc/join_tray", body, &trayID, nil); err != nil {
		return "", err
	}
	return trayID, nil
}

func (b *supabaseBackend) UpdateTrayName(ctx context.Context, sess Session, trayID, name string) error {
	return ErrNotImplemented
}

func (b *supabaseBackend) DeleteTray(ctx context.Context, sess Session, trayID string) error {
	return ErrNotImplemented
}

func (b *supabaseBackend) SetTrayInviteToken(ctx context.Context, sess Session, trayID string, inviteToken *string) error {
	return ErrNotImplemented
}

func (b *supabaseBackend) ListTrayMembers(ctx context.Context, sess Session, trayID string) ([]TrayMember, error) {
	return nil, ErrNotImplemented
}

func (b *supabaseBackend) RemoveTrayMember(ctx context.Context, sess Session, trayID, userID string) error {
	return ErrNotImplemented
}

func (b *supabaseBackend) LeaveTray(ctx context.Context, sess Session, trayID string) error {
	return ErrNotImplemented
}

func (b *supabaseBackend) AddItem(ctx context.Context, sess Session, trayID, title string, dueDate *string) (*Item, error) {
	return nil, ErrNotImplemented
}

func (b *supabaseBackend) ListItems(ctx context.Context, sess Session, q ListItemsQuery) ([]Item, error) {
	return nil, ErrNotImplemented
}

func (b *supabaseBackend) ListOutbox(ctx context.Context, sess Session) ([]Item, error) {
	return nil, ErrNotImplemented
}

func (b *supabaseBackend) UpdateItem(ctx context.Context, sess Session, itemID string, patch ItemPatch) error {
	return ErrNotImplemented
}

func (b *supabaseBackend) DeleteItem(ctx context.Context, sess Session, itemID string) error {
	return ErrNotImplemented
}

// --- HTTP + JSON ---

func hdrPreferRepresentation() http.Header {
	h := http.Header{}
	h.Set("Prefer", "return=representation")
	return h
}

func (b *supabaseBackend) request(ctx context.Context, sess Session, method, path string, body any, extra http.Header) ([]byte, error) {
	var rdr io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("backend: encode body: %w", err)
		}
		rdr = bytes.NewReader(buf)
	}
	req, err := b.client.NewRequest(ctx, method, path, rdr, sess.AccessToken)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	for k, vv := range extra {
		for _, v := range vv {
			req.Header.Add(k, v)
		}
	}
	resp, err := b.client.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("backend: request: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("backend: %s %s: %s: %s", method, path, resp.Status, strings.TrimSpace(string(raw)))
	}
	return raw, nil
}

func (b *supabaseBackend) doJSON(ctx context.Context, sess Session, method, path string, body any, out any, extra http.Header) error {
	raw, err := b.request(ctx, sess, method, path, body, extra)
	if err != nil {
		return err
	}
	if out == nil || len(raw) == 0 {
		return nil
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("backend: decode response: %w", err)
	}
	return nil
}

// PostgREST may return one row as an object or as a single-element array.
func parseCreatedTray(raw []byte) (*trayJSON, error) {
	var rows []trayJSON
	if err := json.Unmarshal(raw, &rows); err == nil && len(rows) > 0 {
		return &rows[0], nil
	}
	var one trayJSON
	if err := json.Unmarshal(raw, &one); err == nil && one.ID != "" {
		return &one, nil
	}
	return nil, fmt.Errorf("backend: parse create tray response: %s", strings.TrimSpace(string(raw)))
}

type trayJSON struct {
	ID          string  `json:"id"`
	OwnerID     string  `json:"owner_id"`
	Name        string  `json:"name"`
	InviteToken *string `json:"invite_token"`
	CreatedAt   string  `json:"created_at"`
}

func trayFromJSON(r trayJSON) (*Tray, error) {
	t, err := parseTime(r.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("backend: tray created_at: %w", err)
	}
	return &Tray{
		ID:          r.ID,
		OwnerID:     r.OwnerID,
		Name:        r.Name,
		InviteToken: r.InviteToken,
		CreatedAt:   t,
	}, nil
}

func parseTime(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, fmt.Errorf("empty time")
	}
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		t, err = time.Parse(time.RFC3339, s)
	}
	return t, err
}
