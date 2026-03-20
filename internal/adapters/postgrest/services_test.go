package postgrest

import (
	"testing"

	"github.com/stretchr/testify/require"
	supabasehttp "github.com/tjdsneto/tray-cli/internal/supabase"
)

func TestNewServices(t *testing.T) {
	c, err := supabasehttp.NewClient("https://x.supabase.co", "anon", nil)
	require.NoError(t, err)
	sv := NewServices(c)
	require.NotNil(t, sv.Trays)
	require.NotNil(t, sv.Items)
}

func TestDial(t *testing.T) {
	sv, err := Dial("https://x.supabase.co", "k", nil)
	require.NoError(t, err)
	require.NotNil(t, sv.Trays)
}
