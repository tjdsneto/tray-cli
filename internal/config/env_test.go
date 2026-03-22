package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOAuthProvider_FromEnv(t *testing.T) {
	t.Setenv(EnvOAuthProvider, "google")
	require.Equal(t, "google", OAuthProvider())
}

func TestDevOAuthHintsEnabled_embedded(t *testing.T) {
	old := EmbeddedDevOAuthHints
	t.Cleanup(func() { EmbeddedDevOAuthHints = old })
	EmbeddedDevOAuthHints = ""
	require.False(t, DevOAuthHintsEnabled())
	EmbeddedDevOAuthHints = "1"
	require.True(t, DevOAuthHintsEnabled())
}

func TestDebug(t *testing.T) {
	t.Setenv(EnvDebug, "")
	require.False(t, Debug())
	t.Setenv(EnvDebug, "1")
	require.True(t, Debug())
	t.Setenv(EnvDebug, "0")
	require.False(t, Debug())
}

func TestSupabaseURL_PrefersEnvOverEmbedded(t *testing.T) {
	oldURL, oldKey := EmbeddedSupabaseURL, EmbeddedSupabaseAnonKey
	t.Cleanup(func() {
		EmbeddedSupabaseURL, EmbeddedSupabaseAnonKey = oldURL, oldKey
	})
	EmbeddedSupabaseURL = "https://embed.example"
	EmbeddedSupabaseAnonKey = "embed-key"

	t.Setenv(EnvSupabaseURL, "https://env.example")
	t.Setenv(EnvSupabaseAnonKey, "")
	require.Equal(t, "https://env.example", SupabaseURL())
	require.Equal(t, "embed-key", SupabaseAnonKey())

	t.Setenv(EnvSupabaseURL, "")
	t.Setenv(EnvSupabaseAnonKey, "")
	require.Equal(t, "https://embed.example", SupabaseURL())
	require.Equal(t, "embed-key", SupabaseAnonKey())
}
