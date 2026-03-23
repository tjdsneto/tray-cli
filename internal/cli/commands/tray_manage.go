package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tjdsneto/tray-cli/internal/cli/trayref"
)

func cmdRename() *cobra.Command {
	return &cobra.Command{
		Use:   "rename <tray> <new-name>",
		Short: "Rename a tray you own",
		Args:  cobra.ExactArgs(2),
		RunE:  runRename,
	}
}

func runRename(cmd *cobra.Command, args []string) error {
	svcs, sess, err := cmdDeps.RequireAuth()
	if err != nil {
		return err
	}
	oldRef := strings.TrimSpace(args[0])
	newName := strings.TrimSpace(args[1])
	if newName == "" {
		return fmt.Errorf("choose a non-empty new name")
	}
	trays, err := svcs.Trays.ListMine(cmd.Context(), sess)
	if err != nil {
		return err
	}
	tid, err := trayref.ResolveTrayRef(cmd.Context(), svcs, sess, oldRef, cmdDeps.RemoteAliases())
	if err != nil {
		return err
	}
	tray, ok := trayref.TrayByID(trays, tid)
	if !ok {
		return fmt.Errorf("tray not found — run `tray ls`")
	}
	if tray.OwnerID != sess.UserID {
		return fmt.Errorf("only the owner can rename %q", tray.Name)
	}
	if err := svcs.Trays.UpdateName(cmd.Context(), sess, tid, newName); err != nil {
		return err
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Renamed tray %q → %q.\n", tray.Name, newName)
	return err
}

func cmdDeleteTray() *cobra.Command {
	return &cobra.Command{
		Use:   "delete-tray <tray>",
		Short: "Permanently delete a tray you own (and its items)",
		Args:  cobra.ExactArgs(1),
		RunE:  runDeleteTray,
	}
}

func runDeleteTray(cmd *cobra.Command, args []string) error {
	svcs, sess, err := cmdDeps.RequireAuth()
	if err != nil {
		return err
	}
	ref := strings.TrimSpace(args[0])
	trays, err := svcs.Trays.ListMine(cmd.Context(), sess)
	if err != nil {
		return err
	}
	tid, err := trayref.ResolveTrayRef(cmd.Context(), svcs, sess, ref, cmdDeps.RemoteAliases())
	if err != nil {
		return err
	}
	tray, ok := trayref.TrayByID(trays, tid)
	if !ok {
		return fmt.Errorf("tray not found — run `tray ls`")
	}
	if tray.OwnerID != sess.UserID {
		return fmt.Errorf("only the owner can delete %q", tray.Name)
	}
	if err := svcs.Trays.Delete(cmd.Context(), sess, tid); err != nil {
		return err
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Deleted tray %q.\n", tray.Name)
	return err
}
