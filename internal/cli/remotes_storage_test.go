package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadSaveRemotes_roundTrip(t *testing.T) {
	dir := t.TempDir()
	f := remotesFile{Aliases: map[string]string{"boss": "00000000-0000-0000-0000-000000000001"}}
	require.NoError(t, saveRemotes(dir, f))

	got, err := loadRemotes(dir)
	require.NoError(t, err)
	require.Equal(t, "00000000-0000-0000-0000-000000000001", got.Aliases["boss"])
}

func TestLoadRemotes_missingFile(t *testing.T) {
	f, err := loadRemotes(t.TempDir())
	require.NoError(t, err)
	require.Empty(t, f.Aliases)
}

func TestLoadRemotes_invalidJSON(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, remotesFileName)
	require.NoError(t, os.WriteFile(p, []byte("{"), 0o600))
	_, err := loadRemotes(dir)
	require.Error(t, err)
}

func TestLoadRemoteAliases_normalizes(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, saveRemotes(dir, remotesFile{Aliases: map[string]string{
		" Boss ": "uuid-1",
	}}))
	m := loadRemoteAliases(dir)
	require.Equal(t, "uuid-1", m["boss"])
}
