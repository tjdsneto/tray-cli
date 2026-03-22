package cli

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUserFacingError_nil(t *testing.T) {
	require.Equal(t, "", UserFacingError(nil))
}

func TestUserFacingError_unknownCommand(t *testing.T) {
	s := UserFacingError(errors.New("unknown command foo"))
	require.Contains(t, s, "That isn't a tray command")
	require.Contains(t, s, "unknown command foo")
}

func TestUserFacingError_unknownFlag(t *testing.T) {
	s := UserFacingError(errors.New("unknown flag: --nope"))
	require.Contains(t, s, "That option isn't recognized")
}

func TestUserFacingError_unknownShorthand(t *testing.T) {
	s := UserFacingError(errors.New("unknown shorthand flag: 'x' in -x"))
	require.Contains(t, s, "That option isn't recognized")
}

func TestUserFacingError_passthrough(t *testing.T) {
	s := UserFacingError(errors.New("something else broke"))
	require.Equal(t, "something else broke", s)
}
