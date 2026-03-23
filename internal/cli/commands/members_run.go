package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tjdsneto/tray-cli/internal/cli/trayref"
	"github.com/tjdsneto/tray-cli/internal/output"
)

func cmdMembers() *cobra.Command {
	return &cobra.Command{
		Use:   "members <tray>",
		Short: "List members of a tray (full list for owner; members see only themselves)",
		Args:  cobra.ExactArgs(1),
		RunE:  runMembers,
	}
}

func runMembers(cmd *cobra.Command, args []string) error {
	svcs, sess, err := cmdDeps.RequireAuth()
	if err != nil {
		return err
	}
	name := strings.TrimSpace(args[0])
	tid, err := trayref.ResolveTrayRef(cmd.Context(), svcs, sess, name, cmdDeps.RemoteAliases())
	if err != nil {
		return err
	}
	members, err := svcs.Trays.ListMembers(cmd.Context(), sess, tid)
	if err != nil {
		return err
	}
	format, err := output.FormatFromCmd(cmd)
	if err != nil {
		return err
	}
	return output.WriteTrayMembers(cmd.OutOrStdout(), name, members, format)
}

func cmdRevoke() *cobra.Command {
	return &cobra.Command{
		Use:   "revoke <tray> <user-id>",
		Short: "Remove a member from a tray (owner only)",
		Args:  cobra.ExactArgs(2),
		RunE:  runRevoke,
	}
}

func runRevoke(cmd *cobra.Command, args []string) error {
	svcs, sess, err := cmdDeps.RequireAuth()
	if err != nil {
		return err
	}
	trayName := strings.TrimSpace(args[0])
	userID := strings.TrimSpace(args[1])
	if userID == "" {
		return fmt.Errorf("pass the member user id (UUID)")
	}
	trays, err := svcs.Trays.ListMine(cmd.Context(), sess)
	if err != nil {
		return err
	}
	tid, err := trayref.ResolveTrayRef(cmd.Context(), svcs, sess, trayName, cmdDeps.RemoteAliases())
	if err != nil {
		return err
	}
	tray, ok := trayref.TrayByID(trays, tid)
	if !ok {
		return fmt.Errorf("tray not found — run `tray ls`")
	}
	if tray.OwnerID != sess.UserID {
		return fmt.Errorf("only the owner can remove members from %q", tray.Name)
	}
	if err := svcs.Trays.RemoveMember(cmd.Context(), sess, tid, userID); err != nil {
		return err
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Removed member %s from tray %q.\n", userID, tray.Name)
	return err
}

func cmdLeave() *cobra.Command {
	return &cobra.Command{
		Use:   "leave <tray>",
		Short: "Leave a tray you joined (tray owners: use delete-tray instead)",
		Args:  cobra.ExactArgs(1),
		RunE:  runLeave,
	}
}

func runLeave(cmd *cobra.Command, args []string) error {
	svcs, sess, err := cmdDeps.RequireAuth()
	if err != nil {
		return err
	}
	trayName := strings.TrimSpace(args[0])
	trays, err := svcs.Trays.ListMine(cmd.Context(), sess)
	if err != nil {
		return err
	}
	tid, err := trayref.ResolveTrayRef(cmd.Context(), svcs, sess, trayName, cmdDeps.RemoteAliases())
	if err != nil {
		return err
	}
	tray, ok := trayref.TrayByID(trays, tid)
	if !ok {
		return fmt.Errorf("tray not found — run `tray ls`")
	}
	if tray.OwnerID == sess.UserID {
		return fmt.Errorf("the tray owner is not a member row — use `tray delete-tray` to remove the tray")
	}
	if err := svcs.Trays.Leave(cmd.Context(), sess, tid); err != nil {
		return err
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Left tray %q.\n", tray.Name)
	return err
}
