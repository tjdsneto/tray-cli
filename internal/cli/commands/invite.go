package commands

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tjdsneto/tray-cli/internal/cli/trayref"
	"github.com/tjdsneto/tray-cli/internal/domain"
	"github.com/tjdsneto/tray-cli/internal/output"
)

func cmdInvite() *cobra.Command {
	c := &cobra.Command{
		Use:   "invite <tray-name>",
		Short: "Show or create the invite token for a tray you own",
		Long: `Only the tray owner can set an invite. If the tray has no token yet, one is created.

Others join with: tray join <token> (or a link that contains the token).`,
		Args: cobra.ExactArgs(1),
		RunE: runInvite,
	}
	c.Flags().Bool("rotate", false, "create a new token and invalidate the previous one")
	return c
}

func runInvite(cmd *cobra.Command, args []string) error {
	rotate, err := cmd.Flags().GetBool("rotate")
	if err != nil {
		return err
	}
	return runInviteCore(cmd, args, rotate)
}

func runInviteCore(cmd *cobra.Command, args []string, rotate bool) error {
	name := strings.TrimSpace(args[0])
	if name == "" {
		return fmt.Errorf("which tray should we invite to? — example: `tray invite inbox`")
	}

	svcs, sess, err := cmdDeps.RequireAuth()
	if err != nil {
		return err
	}
	if strings.TrimSpace(sess.UserID) == "" {
		return fmt.Errorf("missing user id in session — run `tray login` again")
	}

	trays, err := svcs.Trays.ListMine(cmd.Context(), sess)
	if err != nil {
		return err
	}
	tid, err := trayref.ResolveTrayRef(cmd.Context(), svcs, sess, name, cmdDeps.RemoteAliases())
	if err != nil {
		return err
	}
	tray, ok := trayref.TrayByID(trays, tid)
	if !ok {
		return fmt.Errorf("tray not found — run `tray ls`")
	}
	if tray.OwnerID != sess.UserID {
		return fmt.Errorf("only the owner can share invites for %q", tray.Name)
	}

	token, err := inviteTokenForTray(cmd.Context(), svcs, sess, tray, rotate)
	if err != nil {
		return err
	}

	format, err := output.FormatFromCmd(cmd)
	if err != nil {
		return err
	}
	return output.WriteInvite(cmd.OutOrStdout(), tray.Name, token, format)
}

func inviteTokenForTray(ctx context.Context, svcs domain.Services, sess domain.Session, tray domain.Tray, rotate bool) (string, error) {
	if !rotate && tray.InviteToken != nil && strings.TrimSpace(*tray.InviteToken) != "" {
		return strings.TrimSpace(*tray.InviteToken), nil
	}
	tok, err := randomInviteToken()
	if err != nil {
		return "", err
	}
	if err := svcs.Trays.SetInviteToken(ctx, sess, tray.ID, &tok); err != nil {
		return "", err
	}
	return tok, nil
}

func randomInviteToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
