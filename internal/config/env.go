package config

import (
	"os"
	"strings"
)

const (
	EnvSupabaseURL      = "TRAY_SUPABASE_URL"
	EnvSupabaseAnonKey  = "TRAY_SUPABASE_ANON_KEY"
	EnvOAuthProvider    = "TRAY_OAUTH_PROVIDER"
	EnvDebug            = "TRAY_DEBUG"
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

// OAuthProvider returns TRAY_OAUTH_PROVIDER (optional default for tray login --provider).
func OAuthProvider() string {
	return strings.TrimSpace(os.Getenv(EnvOAuthProvider))
}

// Debug is true when TRAY_DEBUG=1 (verbose API errors and diagnostics).
func Debug() bool {
	return strings.TrimSpace(os.Getenv(EnvDebug)) == "1"
}
