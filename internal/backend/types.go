package backend

import "time"

// Session is an authenticated user session (tokens come from login; UserID must be set when known).
type Session struct {
	AccessToken string
	// UserID is the auth provider subject (e.g. Supabase auth.users id). Required for some writes.
	UserID string
}

// Tray is a named inbox owned by one user.
type Tray struct {
	ID          string
	OwnerID     string
	Name        string
	InviteToken *string
	CreatedAt   time.Time
}

// Item is a line on a tray.
type Item struct {
	ID                 string
	TrayID             string
	SourceUserID       string
	Title              string
	Status             string
	DueDate            *string
	SnoozeUntil        *time.Time
	DeclineReason      *string
	CompletionMessage  *string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// TrayMember is a user allowed to add (and read) items on someone else's tray.
type TrayMember struct {
	ID         string
	TrayID     string
	UserID     string
	JoinedAt   time.Time
	InvitedVia *string
}

// ListItemsQuery filters items for a tray (all optional zero values mean "no filter").
type ListItemsQuery struct {
	TrayID        string
	Status        string
	UpdatedAfter  *time.Time
	OrderCreated  string // "asc" | "desc"
}
