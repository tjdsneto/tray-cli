package auth

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewCodeVerifier(t *testing.T) {
	v, c, err := NewCodeVerifier()
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(v), 43)
	require.NotEmpty(t, c)
}
