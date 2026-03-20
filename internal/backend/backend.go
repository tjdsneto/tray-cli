package backend

import (
	"context"
	"errors"
	"time"
)

// ErrNotImplemented is returned by backend operations not wired up yet.
var ErrNotImplemented = errors.New("backend: not implemented yet")

// Backend is the data + auth API the CLI needs. Implementations hide REST/SQL/Firestore details.
type Backend interface {
	// --- Trays ---
	CreateTray(ctx context.Context, sess Session, name string, inviteToken *string) (*Tray, error)
	ListMyTrays(ctx context.Context, sess Session) ([]Tray, error)
	UpdateTrayName(ctx context.Context, sess Session, trayID, name string) error
	DeleteTray(ctx context.Context, sess Session, trayID string) error
	SetTrayInviteToken(ctx context.Context, sess Session, trayID string, inviteToken *string) error

	// JoinTray adds the current user as a member using a share invite token (Model B).
	JoinTray(ctx context.Context, sess Session, inviteToken string) (trayID string, err error)

	ListTrayMembers(ctx context.Context, sess Session, trayID string) ([]TrayMember, error)
	RemoveTrayMember(ctx context.Context, sess Session, trayID, userID string) error
	LeaveTray(ctx context.Context, sess Session, trayID string) error

	// --- Items ---
	AddItem(ctx context.Context, sess Session, trayID, title string, dueDate *string) (*Item, error)
	ListItems(ctx context.Context, sess Session, q ListItemsQuery) ([]Item, error)
	ListOutbox(ctx context.Context, sess Session) ([]Item, error)
	UpdateItem(ctx context.Context, sess Session, itemID string, patch ItemPatch) error
	DeleteItem(ctx context.Context, sess Session, itemID string) error
}

// ItemPatch is owner-only triage updates (partial).
type ItemPatch struct {
	Status              *string
	DeclineReason       *string
	CompletionMessage   *string
	SnoozeUntil         *time.Time
	DueDate             *string
}
