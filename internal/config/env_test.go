package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

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
