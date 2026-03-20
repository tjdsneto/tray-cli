package config

import "os"

const (
	EnvSupabaseURL    = "TRAY_SUPABASE_URL"
	EnvSupabaseAnonKey = "TRAY_SUPABASE_ANON_KEY"
)

// SupabaseURL returns TRAY_SUPABASE_URL (may be empty).
func SupabaseURL() string {
	return os.Getenv(EnvSupabaseURL)
}

// SupabaseAnonKey returns TRAY_SUPABASE_ANON_KEY (may be empty).
func SupabaseAnonKey() string {
	return os.Getenv(EnvSupabaseAnonKey)
}
