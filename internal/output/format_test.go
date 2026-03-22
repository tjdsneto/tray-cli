package output

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestFormat_String(t *testing.T) {
	require.Equal(t, "human", FormatTable.String())
	require.Equal(t, "json", FormatJSON.String())
	require.Equal(t, "markdown", FormatMarkdown.String())
	require.Equal(t, "human", Format(99).String())
}

func TestParse(t *testing.T) {
	for _, tt := range []struct {
		in   string
		want Format
	}{
		{"human", FormatTable},
		{"table", FormatTable},
		{"TABLE", FormatTable},
		{"json", FormatJSON},
		{"machine", FormatJSON},
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

func TestFormatFromCmd_defaultHuman(t *testing.T) {
	root := &cobra.Command{Use: "tray"}
	RegisterPersistentFlags(root.PersistentFlags())
	sub := &cobra.Command{Use: "ls"}
	root.AddCommand(sub)

	require.NoError(t, root.ParseFlags(nil))
	f, err := FormatFromCmd(sub)
	require.NoError(t, err)
	require.Equal(t, FormatTable, f)
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

func TestFormatFromCmd_formatJson(t *testing.T) {
	root := &cobra.Command{Use: "tray"}
	RegisterPersistentFlags(root.PersistentFlags())
	sub := &cobra.Command{Use: "ls"}
	root.AddCommand(sub)

	require.NoError(t, root.ParseFlags([]string{"--format", "json"}))
	f, err := FormatFromCmd(sub)
	require.NoError(t, err)
	require.Equal(t, FormatJSON, f)
}

func TestFormatFromCmd_machineAlias(t *testing.T) {
	root := &cobra.Command{Use: "tray"}
	RegisterPersistentFlags(root.PersistentFlags())
	sub := &cobra.Command{Use: "ls"}
	root.AddCommand(sub)

	require.NoError(t, root.ParseFlags([]string{"--format", "machine"}))
	f, err := FormatFromCmd(sub)
	require.NoError(t, err)
	require.Equal(t, FormatJSON, f)
}

func TestFormatFromCmd_deprecatedOutput(t *testing.T) {
	root := &cobra.Command{Use: "tray"}
	RegisterPersistentFlags(root.PersistentFlags())
	sub := &cobra.Command{Use: "ls"}
	root.AddCommand(sub)

	require.NoError(t, root.ParseFlags([]string{"-o", "json"}))
	f, err := FormatFromCmd(sub)
	require.NoError(t, err)
	require.Equal(t, FormatJSON, f)
}

func TestFormatFromFlags_jsonShorthand(t *testing.T) {
	f, err := FormatFromFlags(true, "human", false, false)
	require.NoError(t, err)
	require.Equal(t, FormatJSON, f)
}

func TestFormatFromFlags_jsonWithExplicitFormat(t *testing.T) {
	f, err := FormatFromFlags(true, "json", true, false)
	require.NoError(t, err)
	require.Equal(t, FormatJSON, f)
}

func TestFormatFromFlags_jsonConflictsWithMarkdown(t *testing.T) {
	_, err := FormatFromFlags(true, "markdown", true, false)
	require.Error(t, err)
}

func TestFormatFromCmd_jsonConflictsWithFormat(t *testing.T) {
	root := &cobra.Command{Use: "tray"}
	RegisterPersistentFlags(root.PersistentFlags())
	sub := &cobra.Command{Use: "ls"}
	root.AddCommand(sub)

	require.NoError(t, root.ParseFlags([]string{"--format", "markdown", "--json"}))
	_, err := FormatFromCmd(sub)
	require.Error(t, err)
}
