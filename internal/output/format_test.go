package output

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	for _, tt := range []struct {
		in   string
		want Format
	}{
		{"table", FormatTable},
		{"TABLE", FormatTable},
		{"json", FormatJSON},
		{"markdown", FormatMarkdown},
		{"md", FormatMarkdown},
	} {
		got, err := Parse(tt.in)
		require.NoError(t, err)
		require.Equal(t, tt.want, got)
	}
	_, err := Parse("xml")
	require.Error(t, err)
}

func TestFormatFromCmd_jsonShorthand(t *testing.T) {
	root := &cobra.Command{Use: "tray"}
	RegisterPersistentFlags(root.PersistentFlags())
	sub := &cobra.Command{Use: "ls"}
	root.AddCommand(sub)

	require.NoError(t, root.ParseFlags([]string{"--json"}))
	f, err := FormatFromCmd(sub)
	require.NoError(t, err)
	require.Equal(t, FormatJSON, f)
}

func TestFormatFromCmd_jsonConflictsWithOutput(t *testing.T) {
	root := &cobra.Command{Use: "tray"}
	RegisterPersistentFlags(root.PersistentFlags())
	sub := &cobra.Command{Use: "ls"}
	root.AddCommand(sub)

	require.NoError(t, root.ParseFlags([]string{"-o", "markdown", "--json"}))
	_, err := FormatFromCmd(sub)
	require.Error(t, err)
}
