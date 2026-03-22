package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// WriteInvite prints invite token and how to join (human / json / markdown).
func WriteInvite(w io.Writer, trayName, token string, f Format) error {
	switch f {
	case FormatJSON:
		type out struct {
			Tray         string `json:"tray"`
			InviteToken  string `json:"invite_token"`
			JoinExample  string `json:"join_example"`
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(out{
			Tray:        trayName,
			InviteToken: token,
			JoinExample: "tray join " + token,
		})
	case FormatMarkdown:
		tok := strings.ReplaceAll(strings.ReplaceAll(token, "|", "\\|"), "\n", " ")
		name := strings.ReplaceAll(strings.ReplaceAll(trayName, "|", "\\|"), "\n", " ")
		_, err := fmt.Fprintf(w, "| | |\n|--|--|\n| tray | %s |\n| invite_token | `%s` |\n| join | `tray join %s` |\n", name, tok, tok)
		return err
	default:
		_, err := fmt.Fprintf(w, "Invite for tray %q\n\n  Token:\n    %s\n\n  Others can join with:\n    tray join %s\n", trayName, token, token)
		return err
	}
}
