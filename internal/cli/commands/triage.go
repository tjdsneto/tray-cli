package commands

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/tjdsneto/tray-cli/internal/cli/triageui"
	"github.com/tjdsneto/tray-cli/internal/cli/trayref"
	"github.com/tjdsneto/tray-cli/internal/domain"
)

func cmdTriage() *cobra.Command {
	return &cobra.Command{
		Use:   "triage [tray]",
		Short: "Interactively triage pending items (TTY)",
		Long: strings.TrimSpace(`
Shows pending items in a full-screen terminal UI. Move with ↑/↓, then:
  a  accept   ·  d  decline (reason)   ·  c  complete (note)   ·  r  archive   ·  q  quit

Decline and complete open a short inline prompt (Enter to confirm, Esc to cancel).
For a non-interactive list, use: tray review`),
		Args: cobra.RangeArgs(0, 1),
		RunE: runTriage,
	}
}

func runTriage(cmd *cobra.Command, args []string) error {
	if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
		return fmt.Errorf("triage requires an interactive terminal")
	}

	svcs, sess, err := cmdDeps.RequireAuth()
	if err != nil {
		return err
	}

	q := domain.ListItemsQuery{Status: "pending", OrderCreated: "desc"}
	if len(args) == 1 {
		tid, err := trayref.ResolveTrayRef(cmd.Context(), svcs, sess, strings.TrimSpace(args[0]), cmdDeps.RemoteAliases())
		if err != nil {
			return err
		}
		q.TrayID = tid
	}

	items, err := svcs.Items.List(cmd.Context(), sess, q)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		_, err := fmt.Fprintln(cmd.OutOrStdout(), "No pending items.")
		return err
	}

	trays, err := svcs.Trays.ListMine(cmd.Context(), sess)
	if err != nil {
		return err
	}
	m := trayref.TrayNameMap(trays)
	by := profileDisplayMap(cmd.Context(), sess, svcs, sourceUserIDsFromItems(items))

	mod := triageui.New(cmd.Context(), svcs, sess, items, m, by)
	p := tea.NewProgram(&mod, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
