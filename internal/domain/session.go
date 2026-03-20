package domain

// Session is an authenticated user session (tokens come from login; UserID must be set when known).
type Session struct {
	AccessToken string
	// UserID is the auth provider subject (e.g. Supabase auth.users id). Required for some writes.
	UserID string
}
