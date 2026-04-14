package domain

import (
	"context"
	"time"
)

// Item is a line on a tray.
type Item struct {
	ID                string
	TrayID            string
	SortOrder         int
	SourceUserID      string
	Title             string
	Status            string
	DueDate           *string
	SnoozeUntil       *time.Time
	DeclineReason     *string
	CompletionMessage *string
	AcceptedAt        *time.Time
	DeclinedAt        *time.Time
	CompletedAt       *time.Time
	ArchivedAt        *time.Time
	SnoozedAt         *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// ListItemsQuery filters items for a tray (all optional zero values mean "no filter").
type ListItemsQuery struct {
	ItemID       string // optional: single item by id
	TrayID       string
	// TrayIDIn limits to items on any of these trays (PostgREST in.); empty means no filter.
	TrayIDIn []string
	Status       string
	UpdatedAfter *time.Time
	// OrderCreated: "asc" | "desc" sorts by created_at only. Empty uses manual tray order (sort_order, then created_at).
	OrderCreated string
}

// ItemPatch is owner-only triage updates (partial).
type ItemPatch struct {
	Status            *string
	SortOrder         *int
	DeclineReason     *string
	CompletionMessage *string
	SnoozeUntil       *time.Time
	DueDate           *string
}

// ItemService is item use-cases on trays the user can access.
type ItemService interface {
	Add(ctx context.Context, sess Session, trayID, title string, dueDate *string) (*Item, error)
	List(ctx context.Context, sess Session, q ListItemsQuery) ([]Item, error)
	ListOutbox(ctx context.Context, sess Session) ([]Item, error)
	Update(ctx context.Context, sess Session, itemID string, patch ItemPatch) error
	Delete(ctx context.Context, sess Session, itemID string) error
	// MoveUp / MoveDown swap sort_order with the adjacent item on the same tray (owner-only).
	MoveUp(ctx context.Context, sess Session, itemID string) error
	MoveDown(ctx context.Context, sess Session, itemID string) error
}
