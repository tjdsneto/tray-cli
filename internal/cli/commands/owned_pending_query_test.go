package commands

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOptionalTrayRefArg(t *testing.T) {
	t.Parallel()
	require.Equal(t, "", optionalTrayRefArg(nil))
	require.Equal(t, "", optionalTrayRefArg([]string{}))
	require.Equal(t, "inbox", optionalTrayRefArg([]string{"inbox"}))
}
