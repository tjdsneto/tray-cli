package credentials

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tjdsneto/tray-cli/internal/domain"
)

func TestSaveLoad_roundTrip(t *testing.T) {
	dir := t.TempDir()
	want := File{
		AccessToken: "at",
		UserID:      "uid-1",
	}
	require.NoError(t, Save(dir, want))
	got, err := Load(dir)
	require.NoError(t, err)
	require.Equal(t, want.AccessToken, got.AccessToken)
	require.Equal(t, want.UserID, got.UserID)

	s := got.Session()
	require.Equal(t, domain.Session{AccessToken: "at", UserID: "uid-1"}, s)
	require.Equal(t, filepath.Join(dir, fileName), Path(dir))
}

func TestLoad_notExists(t *testing.T) {
	_, err := Load(t.TempDir())
	require.ErrorIs(t, err, ErrNotLoggedIn)
}
