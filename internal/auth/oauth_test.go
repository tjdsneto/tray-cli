package auth

import (
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuthorizeURL(t *testing.T) {
	got, err := AuthorizeURL(
		"https://abc.supabase.co",
		"github",
		"http://127.0.0.1:54321/callback",
		"test-challenge",
	)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(got, "https://abc.supabase.co/auth/v1/authorize?"))
	u, err := url.Parse(got)
	require.NoError(t, err)
	q := u.Query()
	require.Equal(t, "github", q.Get("provider"))
	require.Equal(t, "http://127.0.0.1:54321/callback", q.Get("redirect_to"))
	require.Equal(t, "test-challenge", q.Get("code_challenge"))
	require.Equal(t, "s256", q.Get("code_challenge_method"))
}
