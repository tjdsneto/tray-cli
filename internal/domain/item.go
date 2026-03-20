package domain

import (
	"context"
	"time"
)

// Item is a line on a tray.
type Item struct {
	ID                string
	TrayID            string
	SourceUserID      string
	Title             string
	Status            string
	DueDate           *string
	SnoozeUntil       *time.Time
	DeclineReason     *string
	CompletionMessage *string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// ListItemsQuery filters items for a tray (all optional zero values mean "no filter").
type ListItemsQuery struct {
	TrayID       string
	Status       string
	UpdatedAfter *time.Time
	OrderCreated string // "asc" | "desc"
}

// ItemPatch is owner-only triage updates (partial).
type ItemPatch struct {
	Status            *string
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
}
