package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tjdsneto/tray-cli/internal/domain"
	"github.com/tjdsneto/tray-cli/internal/output"
)

func cmdCreate() *cobra.Command {
	return &cobra.Command{
		Use:   "create <name>",
		Short: "Create a named tray",
		Args:  cobra.ExactArgs(1),
		RunE:  runCreate,
	}
}

func runCreate(cmd *cobra.Command, args []string) error {
	name := strings.TrimSpace(args[0])
	if name == "" {
		return fmt.Errorf("give your tray a name — for example: `tray create inbox`")
	}
	svcs, sess, err := cmdDeps.RequireAuth()
	if err != nil {
		return err
	}
	tray, err := svcs.Trays.Create(cmd.Context(), sess, name, nil)
	if err != nil {
		return err
	}
	format, err := output.FormatFromCmd(cmd)
	if err != nil {
		return err
	}
	showHints := format == output.FormatTable
	if format == output.FormatTable || format == output.FormatMarkdown {
		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Created tray %q.\n\n", tray.Name); err != nil {
			return err
		}
	}
	return output.WriteTrays(cmd.OutOrStdout(), []domain.Tray{*tray}, format, showHints)
}

func cmdLs() *cobra.Command {
	return &cobra.Command{
		Use:     "ls",
		Aliases: []string{"list-trays"},
		Short:   "List trays you can access (owned and joined)",
		RunE:    runLs,
	}
}

func runLs(cmd *cobra.Command, args []string) error {
	svcs, sess, err := cmdDeps.RequireAuth()
	if err != nil {
		return err
	}
	trays, err := svcs.Trays.ListMine(cmd.Context(), sess)
	if err != nil {
		return err
	}
	format, err := output.FormatFromCmd(cmd)
	if err != nil {
		return err
	}
	showHints := format == output.FormatTable
	return output.WriteTrays(cmd.OutOrStdout(), trays, format, showHints)
}

func cmdJoin() *cobra.Command {
	return &cobra.Command{
		Use:   "join <url-or-token>",
		Short: "Join a tray via invite URL or token",
		Args:  cobra.ExactArgs(1),
		RunE:  runJoin,
	}
}
