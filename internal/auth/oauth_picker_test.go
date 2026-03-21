package auth

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildPickerLinks(t *testing.T) {
	links, err := buildPickerLinks(
		"https://abc.supabase.co",
		"http://127.0.0.1:5555/callback",
		"challenge",
	)
	require.NoError(t, err)
	require.Len(t, links, len(PickerProviders))
	require.Contains(t, links[0].Href, "provider=")
	require.True(t, strings.HasPrefix(links[0].Href, "https://abc.supabase.co/auth/v1/authorize?"))
}
