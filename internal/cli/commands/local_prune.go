package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tjdsneto/tray-cli/internal/cli/agentsession"
	"github.com/tjdsneto/tray-cli/internal/domain"
	"github.com/tjdsneto/tray-cli/internal/localtray"
)

func cmdPrune() *cobra.Command {
	c := &cobra.Command{
		Use:   "prune",
		Short: "Remove empty local trays, or idle remote agent-session trays",
		Long: `Default: deletes index entries (and empty item files) for local directory and branch trays that have no items. Global trays are kept even when empty.

With --remote (requires sign-in): deletes owned remote agent-session:<id> trays that have no open items and have been idle at least --idle (default 7 days). Other remote trays are never pruned.`,
		RunE: runPrune,
	}
	c.Flags().Bool("dry-run", false, "print trays that would be pruned without changing anything")
	c.Flags().Bool("remote", false, "prune idle empty remote agent-session trays you own (requires sign-in)")
	c.Flags().Duration("idle", agentsession.DefaultPruneIdle, "minimum idle age for remote agent-session prune (e.g. 168h)")
	return c
}

func runPrune(cmd *cobra.Command, args []string) error {
	remote, err := cmd.Flags().GetBool("remote")
	if err != nil {
		return err
	}
	if remote {
		return runRemotePrune(cmd)
	}
	return runLocalPrune(cmd)
}

func runLocalPrune(cmd *cobra.Command) error {
	dryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return err
	}
	pruned, err := localStore().PruneEmpty(dryRun)
	if err != nil {
		return err
	}
	return printPruneResult(cmd, pruned, dryRun)
}

func runRemotePrune(cmd *cobra.Command) error {
	dryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return err
	}
	idle, err := cmd.Flags().GetDuration("idle")
	if err != nil {
		return err
	}
	svcs, sess, err := cmdDeps.RequireAuth()
	if err != nil {
		return err
	}
	owned, err := svcs.Trays.ListOwned(cmd.Context(), sess)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	var pruned []string
	for _, tray := range owned {
		if !agentsession.IsAgentSessionTrayName(tray.Name) {
			continue
		}
		items, listErr := svcs.Items.List(cmd.Context(), sess, domain.ListItemsQuery{TrayID: tray.ID})
		if listErr != nil {
			return listErr
		}
		if !agentsession.Prunable(tray, items, now, idle) {
			continue
		}
		if !dryRun {
			if delErr := svcs.Trays.Delete(cmd.Context(), sess, tray.ID); delErr != nil {
				return delErr
			}
		}
		pruned = append(pruned, tray.Name)
	}
	return printPruneResult(cmd, pruned, dryRun)
}

func printPruneResult(cmd *cobra.Command, pruned []string, dryRun bool) error {
	if len(pruned) == 0 {
		_, err := fmt.Fprintln(cmd.OutOrStdout(), "Nothing to prune.")
		return err
	}
	verb := "Pruned"
	if dryRun {
		verb = "Would prune"
	}
	for _, id := range pruned {
		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", verb, id); err != nil {
			return err
		}
	}
	return nil
}

func localListTrayIDs(all bool, cwd string, git localtray.GitInfo, trayArg string) ([]string, error) {
	if strings.TrimSpace(trayArg) != "" {
		target, err := localtray.TargetFromRef(trayArg, cwd, git)
		if err != nil {
			return nil, err
		}
		id, err := trayIDFromTarget(target)
		if err != nil {
			return nil, err
		}
		idx, err := loadLocalIndex()
		if err != nil {
			return nil, err
		}
		if _, ok := idx.Find(id); !ok {
			return nil, fmt.Errorf("local tray %q does not exist yet — add an item first", localtray.DisplayLabel(id))
		}
		return []string{id}, nil
	}
	return localTrayIDs(all, cwd, git)
}

func trayIDFromTarget(t localtray.AddTarget) (string, error) {
	switch t.Kind {
	case localtray.KindGlobal:
		return localtray.GlobalID(t.Name)
	case localtray.KindDir:
		return localtray.DirID(t.Path)
	case localtray.KindBranch:
		return localtray.BranchID(t.RepoRoot, t.Branch)
	default:
		return "", fmt.Errorf("unknown tray kind %q", t.Kind)
	}
}
