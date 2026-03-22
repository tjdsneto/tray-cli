package output

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Format is how list-style commands render to stdout.
type Format int

const (
	FormatTable Format = iota
	FormatJSON
	FormatMarkdown
)

// String returns the canonical CLI token for f (for help and errors).
func (f Format) String() string {
	switch f {
	case FormatTable:
		return "human"
	case FormatJSON:
		return "json"
	case FormatMarkdown:
		return "markdown"
	default:
		return "human"
	}
}

// Parse normalizes s (e.g. "md" → markdown, "machine" → json).
func Parse(s string) (Format, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "human", "table", "text":
		return FormatTable, nil
	case "json", "machine":
		return FormatJSON, nil
	case "markdown", "md":
		return FormatMarkdown, nil
	default:
		return 0, fmt.Errorf("unknown --format %q (use human, json, machine, markdown, or md)", s)
	}
}

// FormatFromFlags resolves --format / --output / --json without cobra (pure; unit-test friendly).
// formatChanged and outputChanged correspond to pflag "Changed" for --format and -o/--output.
func FormatFromFlags(jsonShort bool, formatOutput string, formatChanged, outputChanged bool) (Format, error) {
	if jsonShort {
		explicit := formatChanged || outputChanged
		if explicit {
			p, err := Parse(formatOutput)
			if err != nil {
				return 0, err
			}
			if p != FormatJSON {
				return 0, fmt.Errorf("cannot combine --json with --format %s", formatOutput)
			}
		}
		return FormatJSON, nil
	}
	return Parse(formatOutput)
}

// FormatFromCmd reads persistent --format, deprecated -o/--output, and --json from the command root.
func FormatFromCmd(cmd *cobra.Command) (Format, error) {
	root := cmd.Root()
	fs := root.PersistentFlags()

	jsonShort, err := fs.GetBool("json")
	if err != nil {
		return 0, err
	}
	out, err := fs.GetString("format")
	if err != nil {
		return 0, err
	}

	ff := fs.Lookup("format")
	fo := fs.Lookup("output")
	var formatChanged, outputChanged bool
	if ff != nil {
		formatChanged = ff.Changed
	}
	if fo != nil {
		outputChanged = fo.Changed
	}
	return FormatFromFlags(jsonShort, out, formatChanged, outputChanged)
}

// RegisterPersistentFlags adds --format (default human), deprecated -o/--output (same meaning), and --json.
func RegisterPersistentFlags(fs *pflag.FlagSet) {
	v := new(string)
	*v = "human"
	const usage = `how to print: "human" (default, friendly tables and hints), "json" or "machine" (stable for scripts), "markdown" or "md" (paste-friendly tables)`
	fs.StringVar(v, "format", "human", usage)
	fs.StringVarP(v, "output", "o", "human", "deprecated: use --format instead")
	_ = fs.MarkDeprecated("output", "use --format instead")
	fs.Bool("json", false, "shorthand for --format json (machine-readable output)")
}
