package cli

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUserFacingError_unknownCommand(t *testing.T) {
	err := errors.New(`unknown command "foo" for "tray"`)
	out := UserFacingError(err)
	require.Contains(t, out, "isn't a tray command")
	require.Contains(t, out, "unknown command")
}

func TestUserFacingError_passthrough(t *testing.T) {
	err := errors.New("something else")
	require.Equal(t, "something else", UserFacingError(err))
}
