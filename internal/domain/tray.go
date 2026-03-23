package domain

import (
	"context"
	"time"
)

// Tray is a named inbox owned by one user.
type Tray struct {
	ID          string    `json:"id"`
	OwnerID     string    `json:"owner_id"`
	Name        string    `json:"name"`
	InviteToken *string   `json:"invite_token,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	// ItemCount is the number of items visible to the current user (from PostgREST embed).
	ItemCount int `json:"item_count"`
}

// TrayMember is a user allowed to add (and read) items on someone else's tray.
type TrayMember struct {
	ID         string    `json:"id"`
	TrayID     string    `json:"tray_id"`
	UserID     string    `json:"user_id"`
	JoinedAt   time.Time `json:"joined_at"`
	InvitedVia *string   `json:"invited_via,omitempty"`
}

// TrayService is tray + membership use-cases (implementations hide storage details).
type TrayService interface {
	Create(ctx context.Context, sess Session, name string, inviteToken *string) (*Tray, error)
	ListMine(ctx context.Context, sess Session) ([]Tray, error)
	UpdateName(ctx context.Context, sess Session, trayID, name string) error
	Delete(ctx context.Context, sess Session, trayID string) error
	SetInviteToken(ctx context.Context, sess Session, trayID string, inviteToken *string) error

	// Join adds the current user as a member using a share invite token (Model B).
	Join(ctx context.Context, sess Session, inviteToken string) (trayID string, err error)

	ListMembers(ctx context.Context, sess Session, trayID string) ([]TrayMember, error)
	RemoveMember(ctx context.Context, sess Session, trayID, userID string) error
	Leave(ctx context.Context, sess Session, trayID string) error
}
