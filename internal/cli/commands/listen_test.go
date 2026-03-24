package commands

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tjdsneto/tray-cli/internal/domain"
)

func TestCmdListen_Metadata(t *testing.T) {
	t.Parallel()
	c := cmdListen()
	require.Equal(t, "listen [tray]", c.Use)
	require.NotNil(t, c.RunE)
	f := c.Flags().Lookup("interval")
	require.NotNil(t, f)
	require.Equal(t, (10 * time.Second).String(), f.DefValue)
}

func TestUnseenItems_marksAndReturnsNew(t *testing.T) {
	t.Parallel()
	seen := map[string]struct{}{"a": {}}
	items := []domain.Item{{ID: "a"}, {ID: "b"}, {ID: "  "}, {ID: "c"}}
	got := unseenItems(items, seen)
	require.Len(t, got, 2)
	require.Equal(t, "b", got[0].ID)
	require.Equal(t, "c", got[1].ID)
	_, okB := seen["b"]
	_, okC := seen["c"]
	require.True(t, okB)
	require.True(t, okC)
}
