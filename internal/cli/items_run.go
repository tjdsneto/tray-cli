package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tjdsneto/tray-cli/internal/domain"
	"github.com/tjdsneto/tray-cli/internal/output"
)

func cmdAdd() *cobra.Command {
	return &cobra.Command{
		Use:   `add "title" <tray>`,
		Short: "Add an item to a tray",
		Long:  `Creates a pending item. Tray can be a name from tray ls or a tray id.`,
		Args:  cobra.ExactArgs(2),
		RunE:  runAdd,
	}
}

func runAdd(cmd *cobra.Command, args []string) error {
	title := strings.TrimSpace(args[0])
	if title == "" {
		return fmt.Errorf("give the item a title — example: tray add \"Fix login\" inbox")
	}
	trayRef := strings.TrimSpace(args[1])
	svcs, sess, err := requireAuth()
	if err != nil {
		return err
	}
	tid, err := resolveTrayRef(cmd.Context(), svcs, sess, trayRef, trayAliasesOrNil())
	if err != nil {
		return err
	}
	item, err := svcs.Items.Add(cmd.Context(), sess, tid, title, nil)
	if err != nil {
		return err
	}
	format, err := output.FormatFromCmd(cmd)
	if err != nil {
		return err
	}
	trays, err := svcs.Trays.ListMine(cmd.Context(), sess)
	if err != nil {
		return err
	}
	m := trayNameMap(trays)
	return output.WriteItems(cmd.OutOrStdout(), []domain.Item{*item}, m, format)
}

func cmdList() *cobra.Command {
	return &cobra.Command{
		Use:   "list [tray]",
		Short: "List items in a tray (or all visible items)",
		Args:  cobra.RangeArgs(0, 1),
		RunE:  runList,
	}
}

func runList(cmd *cobra.Command, args []string) error {
	svcs, sess, err := requireAuth()
	if err != nil {
		return err
	}
	q := domain.ListItemsQuery{OrderCreated: "desc"}
	if len(args) == 1 {
		tid, err := resolveTrayRef(cmd.Context(), svcs, sess, strings.TrimSpace(args[0]), trayAliasesOrNil())
		if err != nil {
			return err
		}
		q.TrayID = tid
	}
	items, err := svcs.Items.List(cmd.Context(), sess, q)
	if err != nil {
		return err
	}
	format, err := output.FormatFromCmd(cmd)
	if err != nil {
		return err
	}
	trays, err := svcs.Trays.ListMine(cmd.Context(), sess)
	if err != nil {
		return err
	}
	m := trayNameMap(trays)
	return output.WriteItems(cmd.OutOrStdout(), items, m, format)
}

func cmdContributed() *cobra.Command {
	return &cobra.Command{
		Use:   "contributed",
		Short: "List items you added to others' trays",
		RunE:  runContributed,
	}
}

func runContributed(cmd *cobra.Command, args []string) error {
	svcs, sess, err := requireAuth()
	if err != nil {
		return err
	}
	items, err := svcs.Items.ListOutbox(cmd.Context(), sess)
	if err != nil {
		return err
	}
	format, err := output.FormatFromCmd(cmd)
	if err != nil {
		return err
	}
	trays, err := svcs.Trays.ListMine(cmd.Context(), sess)
	if err != nil {
		return err
	}
	m := trayNameMap(trays)
	return output.WriteItems(cmd.OutOrStdout(), items, m, format)
}

// trayAliasesOrNil returns remote aliases when implemented; nil means name/id only.
func trayAliasesOrNil() map[string]string {
	return loadRemoteAliases(ConfigDir())
}
