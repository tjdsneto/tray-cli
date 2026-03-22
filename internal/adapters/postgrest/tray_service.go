package postgrest

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/tjdsneto/tray-cli/internal/domain"
)

type trayService struct {
	pg *client
}

func newTrayService(pg *client) *trayService {
	return &trayService{pg: pg}
}

var _ domain.TrayService = (*trayService)(nil)

func (s *trayService) Create(ctx context.Context, sess domain.Session, name string, inviteToken *string) (*domain.Tray, error) {
	if strings.TrimSpace(sess.UserID) == "" {
		return nil, fmt.Errorf("postgrest: session missing UserID (set after login)")
	}
	body := map[string]any{
		"name":     name,
		"owner_id": sess.UserID,
	}
	if inviteToken != nil {
		body["invite_token"] = *inviteToken
	}
	raw, err := s.pg.request(ctx, sess, http.MethodPost, "/rest/v1/trays", body, hdrPreferRepresentation())
	if err != nil {
		return nil, err
	}
	row, err := parseCreatedTray(raw)
	if err != nil {
		return nil, err
	}
	return trayFromRow(*row)
}

func (s *trayService) ListMine(ctx context.Context, sess domain.Session) ([]domain.Tray, error) {
	q := url.Values{}
	q.Set("select", "id,owner_id,name,invite_token,created_at,items(count)")
	q.Set("order", "name.asc")
	path := "/rest/v1/trays?" + q.Encode()
	var rows []trayRow
	if err := s.pg.doJSON(ctx, sess, http.MethodGet, path, nil, &rows, nil); err != nil {
		return nil, err
	}
	out := make([]domain.Tray, 0, len(rows))
	for _, r := range rows {
		t, err := trayFromRow(r)
		if err != nil {
			return nil, err
		}
		out = append(out, *t)
	}
	return out, nil
}

func (s *trayService) Join(ctx context.Context, sess domain.Session, inviteToken string) (string, error) {
	body := map[string]string{"p_invite_token": strings.TrimSpace(inviteToken)}
	var trayID string
	if err := s.pg.doJSON(ctx, sess, http.MethodPost, "/rest/v1/rpc/join_tray", body, &trayID, nil); err != nil {
		return "", err
	}
	return trayID, nil
}

func (s *trayService) UpdateName(ctx context.Context, sess domain.Session, trayID, name string) error {
	return domain.ErrNotImplemented
}

func (s *trayService) Delete(ctx context.Context, sess domain.Session, trayID string) error {
	return domain.ErrNotImplemented
}

func (s *trayService) SetInviteToken(ctx context.Context, sess domain.Session, trayID string, inviteToken *string) error {
	if strings.TrimSpace(sess.UserID) == "" {
		return fmt.Errorf("postgrest: session missing UserID (set after login)")
	}
	tid := strings.TrimSpace(trayID)
	if tid == "" {
		return fmt.Errorf("postgrest: empty tray id")
	}
	q := url.Values{}
	q.Set("id", "eq."+tid)
	path := "/rest/v1/trays?" + q.Encode()
	var body map[string]any
	if inviteToken != nil {
		body = map[string]any{"invite_token": *inviteToken}
	} else {
		body = map[string]any{"invite_token": nil}
	}
	_, err := s.pg.request(ctx, sess, http.MethodPatch, path, body, nil)
	return err
}

func (s *trayService) ListMembers(ctx context.Context, sess domain.Session, trayID string) ([]domain.TrayMember, error) {
	return nil, domain.ErrNotImplemented
}

func (s *trayService) RemoveMember(ctx context.Context, sess domain.Session, trayID, userID string) error {
	return domain.ErrNotImplemented
}

func (s *trayService) Leave(ctx context.Context, sess domain.Session, trayID string) error {
	return domain.ErrNotImplemented
}
