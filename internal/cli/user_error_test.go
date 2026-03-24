package cli

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tjdsneto/tray-cli/internal/cli/errs"
)

func TestUserFacingError_nil(t *testing.T) {
	require.Equal(t, "", UserFacingError(nil))
}

func TestUserFacingError_missingBackend(t *testing.T) {
	s := UserFacingError(fmt.Errorf("dial: %w", errs.MissingBackendConfig))
	require.Contains(t, s, "isn’t configured")
	require.NotContains(t, s, "run.sh")
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

func TestWriteUserError_noDebug(t *testing.T) {
	var b bytes.Buffer
	WriteUserError(&b, errors.New("boom"), false)
	s := b.String()
	require.Contains(t, s, "tray: boom")
	require.NotContains(t, s, "[debug]")
}

func TestWriteUserError_withDebug(t *testing.T) {
	var b bytes.Buffer
	WriteUserError(&b, errors.New("boom"), true)
	s := b.String()
	require.Contains(t, s, "tray [debug] boom")
	require.Contains(t, s, "tray: boom")
}
