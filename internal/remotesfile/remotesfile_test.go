package remotesfile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadSave_roundTrip(t *testing.T) {
	dir := t.TempDir()
	f := File{Aliases: map[string]string{"boss": "00000000-0000-0000-0000-000000000001"}}
	require.NoError(t, Save(dir, f))

	got, err := Load(dir)
	require.NoError(t, err)
	require.Equal(t, "00000000-0000-0000-0000-000000000001", got.Aliases["boss"])
}

func TestLoad_missingFile(t *testing.T) {
	f, err := Load(t.TempDir())
	require.NoError(t, err)
	require.Empty(t, f.Aliases)
}

func TestLoad_invalidJSON(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, fileName)
	require.NoError(t, os.WriteFile(p, []byte("{"), 0o600))
	_, err := Load(dir)
	require.Error(t, err)
}

func TestAliasesMap_normalizes(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, Save(dir, File{Aliases: map[string]string{
		" Boss ": "uuid-1",
	}}))
	m := AliasesMap(dir)
	require.Equal(t, "uuid-1", m["boss"])
}
