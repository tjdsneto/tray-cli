package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVersionLine(t *testing.T) {
	savedV := Version
	savedC := GitCommit
	t.Cleanup(func() {
		Version = savedV
		GitCommit = savedC
	})

	Version = "dev"
	GitCommit = ""
	require.Equal(t, "dev", VersionLine())

	Version = "v1.2.3"
	GitCommit = ""
	require.Equal(t, "v1.2.3", VersionLine())

	GitCommit = "abc1234"
	require.Equal(t, "v1.2.3 (abc1234)", VersionLine())
}
