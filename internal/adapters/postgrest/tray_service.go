package postgrest

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/tjdsneto/tray-cli/internal/adapters/postgrest/pghttp"
	"github.com/tjdsneto/tray-cli/internal/domain"
)

type trayService struct {
	pg *pghttp.Client
}

func newTrayService(pg *pghttp.Client) *trayService {
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
	raw, err := s.pg.Request(ctx, sess.AccessToken, http.MethodPost, traysCreatePath(), body, pghttp.PreferRepresentation())
	if err != nil {
		return nil, err
	}
	row, err := parseCreatedTray(raw)
	if err != nil {
		return nil, err
	}
	return row.ToDomain()
}

func (s *trayService) ListMine(ctx context.Context, sess domain.Session) ([]domain.Tray, error) {
	path := traysListMinePath()
	var rows []trayRow
	if err := s.pg.DoJSON(ctx, sess.AccessToken, http.MethodGet, path, nil, &rows, nil); err != nil {
		return nil, err
	}
	out := make([]domain.Tray, 0, len(rows))
	for _, r := range rows {
		t, err := r.ToDomain()
		if err != nil {
			return nil, err
		}
		out = append(out, *t)
	}
	return out, nil
}

func (s *trayService) Join(ctx context.Context, sess domain.Session, inviteToken string) (string, error) {
	body := joinTrayRequest{PInviteToken: strings.TrimSpace(inviteToken)}
	var trayID string
	if err := s.pg.DoJSON(ctx, sess.AccessToken, http.MethodPost, joinTrayRPCPath(), body, &trayID, nil); err != nil {
		return "", err
	}
	return trayID, nil
}

func (s *trayService) UpdateName(ctx context.Context, sess domain.Session, trayID, name string) error {
	if strings.TrimSpace(sess.UserID) == "" {
		return fmt.Errorf("postgrest: session missing UserID (set after login)")
	}
	tid := strings.TrimSpace(trayID)
	n := strings.TrimSpace(name)
	if tid == "" || n == "" {
		return fmt.Errorf("postgrest: empty tray id or name")
	}
	path := traysByIDPath(tid)
	body := map[string]any{"name": n}
	_, err := s.pg.Request(ctx, sess.AccessToken, http.MethodPatch, path, body, nil)
	return err
}

func (s *trayService) Delete(ctx context.Context, sess domain.Session, trayID string) error {
	tid := strings.TrimSpace(trayID)
	if tid == "" {
		return fmt.Errorf("postgrest: empty tray id")
	}
	path := traysByIDPath(tid)
	_, err := s.pg.Request(ctx, sess.AccessToken, http.MethodDelete, path, nil, nil)
	return err
}

func (s *trayService) SetInviteToken(ctx context.Context, sess domain.Session, trayID string, inviteToken *string) error {
	if strings.TrimSpace(sess.UserID) == "" {
		return fmt.Errorf("postgrest: session missing UserID (set after login)")
	}
	tid := strings.TrimSpace(trayID)
	if tid == "" {
		return fmt.Errorf("postgrest: empty tray id")
	}
	path := traysByIDPath(tid)
	var body map[string]any
	if inviteToken != nil {
		body = map[string]any{"invite_token": *inviteToken}
	} else {
		body = map[string]any{"invite_token": nil}
	}
	_, err := s.pg.Request(ctx, sess.AccessToken, http.MethodPatch, path, body, nil)
	return err
}

func (s *trayService) ListMembers(ctx context.Context, sess domain.Session, trayID string) ([]domain.TrayMember, error) {
	tid := strings.TrimSpace(trayID)
	if tid == "" {
		return nil, fmt.Errorf("postgrest: empty tray id")
	}
	path := trayMembersListPath(tid)
	raw, err := s.pg.Request(ctx, sess.AccessToken, http.MethodGet, path, nil, nil)
	if err != nil {
		return nil, err
	}
	return parseTrayMemberRows(raw)
}

func (s *trayService) RemoveMember(ctx context.Context, sess domain.Session, trayID, userID string) error {
	path := trayMembersDeletePath(trayID, userID)
	_, err := s.pg.Request(ctx, sess.AccessToken, http.MethodDelete, path, nil, nil)
	return err
}

func (s *trayService) Leave(ctx context.Context, sess domain.Session, trayID string) error {
	if strings.TrimSpace(sess.UserID) == "" {
		return fmt.Errorf("postgrest: session missing UserID (set after login)")
	}
	return s.RemoveMember(ctx, sess, trayID, sess.UserID)
}

const trayMemberSelectColumns = "id,tray_id,user_id,joined_at,invited_via"

func traysCreatePath() string {
	return "/rest/v1/trays"
}

// traysListMinePath is GET /rest/v1/trays with select and order for the current user's trays.
func traysListMinePath() string {
	q := url.Values{}
	q.Set("select", "id,owner_id,name,invite_token,created_at,items(count)")
	q.Set("order", "name.asc")
	return "/rest/v1/trays?" + q.Encode()
}

func joinTrayRPCPath() string {
	return "/rest/v1/rpc/join_tray"
}

func traysByIDPath(trayID string) string {
	q := url.Values{}
	q.Set("id", "eq."+strings.TrimSpace(trayID))
	return "/rest/v1/trays?" + q.Encode()
}

func trayMembersListPath(trayID string) string {
	q := url.Values{}
	q.Set("tray_id", "eq."+strings.TrimSpace(trayID))
	q.Set("select", trayMemberSelectColumns)
	q.Set("order", "joined_at.asc")
	return "/rest/v1/tray_members?" + q.Encode()
}

func trayMembersDeletePath(trayID, userID string) string {
	q := url.Values{}
	q.Set("tray_id", "eq."+strings.TrimSpace(trayID))
	q.Set("user_id", "eq."+strings.TrimSpace(userID))
	return "/rest/v1/tray_members?" + q.Encode()
}
