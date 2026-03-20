package config

import "os"

const (
	EnvSupabaseURL     = "TRAY_SUPABASE_URL"
	EnvSupabaseAnonKey = "TRAY_SUPABASE_ANON_KEY"
)

// Set at link time via -ldflags -X (see build.sh / run.sh and CI).
var (
	EmbeddedSupabaseURL     = ""
	EmbeddedSupabaseAnonKey = ""
)

// SupabaseURL returns TRAY_SUPABASE_URL if set, else the value embedded at build time.
func SupabaseURL() string {
	if v := os.Getenv(EnvSupabaseURL); v != "" {
		return v
	}
	return EmbeddedSupabaseURL
}

// SupabaseAnonKey returns TRAY_SUPABASE_ANON_KEY if set, else the value embedded at build time.
func SupabaseAnonKey() string {
	if v := os.Getenv(EnvSupabaseAnonKey); v != "" {
		return v
	}
	return EmbeddedSupabaseAnonKey
}
