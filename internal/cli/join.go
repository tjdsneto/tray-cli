package cli

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tjdsneto/tray-cli/internal/output"
)

// extractInviteToken accepts a raw invite token or a URL that carries the token in a query
// parameter (token, invite_token) or as the URL fragment.
func extractInviteToken(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	u, err := url.Parse(s)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return s
	}
	if t := u.Query().Get("token"); t != "" {
		return strings.TrimSpace(t)
	}
	if t := u.Query().Get("invite_token"); t != "" {
		return strings.TrimSpace(t)
	}
	if frag := strings.TrimSpace(u.Fragment); frag != "" {
		return frag
	}
	return s
}

func runJoin(cmd *cobra.Command, args []string) error {
	token := extractInviteToken(args[0])
	if token == "" {
		return fmt.Errorf("paste an invite token or a link that contains the token")
	}
	svcs, sess, err := requireAuth()
	if err != nil {
		return err
	}
	trayID, err := svcs.Trays.Join(cmd.Context(), sess, token)
	if err != nil {
		return err
	}

	var name string
	if trays, err := svcs.Trays.ListMine(cmd.Context(), sess); err == nil {
		for i := range trays {
			if trays[i].ID == trayID {
				name = trays[i].Name
				break
			}
		}
	}

	format, err := output.FormatFromCmd(cmd)
	if err != nil {
		return err
	}
	return output.WriteJoin(cmd.OutOrStdout(), trayID, name, format)
}
