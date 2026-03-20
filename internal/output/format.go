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

// String returns the canonical CLI token for f.
func (f Format) String() string {
	switch f {
	case FormatTable:
		return "table"
	case FormatJSON:
		return "json"
	case FormatMarkdown:
		return "markdown"
	default:
		return "table"
	}
}

// Parse normalizes s (e.g. "md" → markdown).
func Parse(s string) (Format, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "table", "text":
		return FormatTable, nil
	case "json":
		return FormatJSON, nil
	case "markdown", "md":
		return FormatMarkdown, nil
	default:
		return 0, fmt.Errorf("unknown output format %q (use table, json, or markdown)", s)
	}
}

// FormatFromCmd reads persistent -o/--output and --json from the command root.
// --json is shorthand for -o json; it conflicts with an explicit -o that is not json.
func FormatFromCmd(cmd *cobra.Command) (Format, error) {
	root := cmd.Root()
	fs := root.PersistentFlags()

	jsonShort, err := fs.GetBool("json")
	if err != nil {
		return 0, err
	}
	out, err := fs.GetString("output")
	if err != nil {
		return 0, err
	}

	if jsonShort {
		if of := fs.Lookup("output"); of != nil && of.Changed {
			o, _ := Parse(out)
			if o != FormatJSON {
				return 0, fmt.Errorf("cannot combine --json with -o %s", out)
			}
		}
		return FormatJSON, nil
	}

	return Parse(out)
}

// RegisterPersistentFlags adds -o/--output and --json to fs (typically the root command).
func RegisterPersistentFlags(fs *pflag.FlagSet) {
	fs.StringP("output", "o", "table", "output format: table (human), json (machines), markdown or md (AI-friendly tables)")
	fs.Bool("json", false, "shorthand for -o json")
}
