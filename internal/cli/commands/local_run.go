package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tjdsneto/tray-cli/internal/localtray"
	"github.com/tjdsneto/tray-cli/internal/output"
)

func localStore() *localtray.Store {
	return localtray.NewStore(cmdDeps.ConfigDir())
}

func workContext() (cwd string, git localtray.GitInfo) {
	cwd = localtray.WorkingCwd()
	git = localtray.DetectGit(cwd)
	return cwd, git
}

func loadLocalIndex() (localtray.Index, error) {
	return localtray.LoadIndex(localtray.LocalDir(cmdDeps.ConfigDir()))
}

func localTrayIDs(all bool, cwd string, git localtray.GitInfo) ([]string, error) {
	idx, err := loadLocalIndex()
	if err != nil {
		return nil, err
	}
	if all {
		return localtray.AllTrayIDs(idx), nil
	}
	return localtray.ContextTrayIDs(idx, cwd, git), nil
}

func runLocalAdd(cmd *cobra.Command, title string, target localtray.AddTarget, create bool) error {
	store := localStore()
	item, _, err := store.AddItem(target, title, create)
	if err != nil {
		return err
	}
	format, err := output.FormatFromCmd(cmd)
	if err != nil {
		return err
	}
	return output.WriteLocalItemAdded(cmd.OutOrStdout(), item, format)
}

func runLocalList(cmd *cobra.Command, trayIDs []string) error {
	store := localStore()
	rows, err := store.ListItemsForTrays(trayIDs)
	if err != nil {
		return err
	}
	format, err := output.FormatFromCmd(cmd)
	if err != nil {
		return err
	}
	return output.WriteLocalItems(cmd.OutOrStdout(), rows, format)
}

func runLocalLs(cmd *cobra.Command, trayIDs []string) error {
	store := localStore()
	summaries, err := store.TraySummaries(trayIDs)
	if err != nil {
		return err
	}
	format, err := output.FormatFromCmd(cmd)
	if err != nil {
		return err
	}
	return output.WriteLocalTrays(cmd.OutOrStdout(), summaries, format)
}

func resolveLocalItemID(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("empty item id")
	}
	store := localStore()
	items, err := store.AllOpenItems()
	if err != nil {
		return "", err
	}
	for _, row := range items {
		if row.Item.ID == raw {
			return raw, nil
		}
	}
	if len(raw) >= 8 {
		return store.ResolveItemPrefix(raw)
	}
	return "", fmt.Errorf("no local item with id %q", raw)
}

func tryLocalComplete(itemID string) (id string, ok bool, err error) {
	id, err = resolveLocalItemID(itemID)
	if err != nil {
		return "", false, nil
	}
	_, err = localStore().CompleteItem(id)
	return id, true, err
}

func tryLocalRemove(itemID string) (id string, ok bool, err error) {
	id, err = resolveLocalItemID(itemID)
	if err != nil {
		return "", false, nil
	}
	return id, true, localStore().RemoveItem(id)
}
