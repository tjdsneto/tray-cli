package commands

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCmdReview_Metadata(t *testing.T) {
	t.Parallel()
	c := cmdReview()
	require.Equal(t, "review [tray]", c.Use)
	require.Equal(t, "Review pending items (owner triage queue)", c.Short)
	require.NotNil(t, c.RunE)
}
