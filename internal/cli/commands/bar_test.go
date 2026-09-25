package commands

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCmdBar_use(t *testing.T) {
	c := cmdBar()
	require.Equal(t, "bar", c.Name())
	require.NotNil(t, c.Flags().Lookup("interval"))
	require.NotNil(t, c.Flags().Lookup("daemon"))
}
