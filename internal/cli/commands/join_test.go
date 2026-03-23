package commands

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractInviteToken(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in, want string
	}{
		{"plain-token", "plain-token"},
		{"  spaced  ", "spaced"},
		{"https://example.com/app?token=abc123", "abc123"},
		{"https://example.com/x?invite_token=xyz&other=1", "xyz"},
		{"https://example.com/#fraggy", "fraggy"},
	}
	for i, tt := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			require.Equal(t, tt.want, extractInviteToken(tt.in))
		})
	}
}
