package commands

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/tjdsneto/tray-cli/internal/cli/trayref"
	"github.com/tjdsneto/tray-cli/internal/output"
	"github.com/tjdsneto/tray-cli/internal/remotesfile"
)

func runRemoteAdd(cmd *cobra.Command, args []string) error {
	alias := strings.TrimSpace(args[0])
	tok := extractInviteToken(args[1])
	if alias == "" {
		return fmt.Errorf("choose a short alias (e.g. `team`)")
	}
	if tok == "" {
		return fmt.Errorf("paste an invite token or link that contains it")
	}
	svcs, sess, err := cmdDeps.RequireAuth()
	if err != nil {
		return err
	}
	trayID, err := svcs.Trays.Join(cmd.Context(), sess, tok)
	if err != nil {
		return err
	}
	f, err := remotesfile.Load(cmdDeps.ConfigDir())
	if err != nil {
		return err
	}
	if f.Aliases == nil {
		f.Aliases = map[string]string{}
	}
	f.Aliases[alias] = trayID
	if err := remotesfile.Save(cmdDeps.ConfigDir(), f); err != nil {
		return err
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Saved alias %q → tray %s (you can use the alias anywhere a tray name goes).\n", alias, trayID)
	return err
}

func runRemoteRename(cmd *cobra.Command, args []string) error {
	curr := strings.TrimSpace(args[0])
	newName := strings.TrimSpace(args[1])
	if curr == "" {
		return fmt.Errorf("give the current remote alias or tray name — example: `tray remote rename work tiago-work`")
	}
	if newName == "" {
		return fmt.Errorf("choose a non-empty new local name")
	}
	if strings.EqualFold(curr, newName) {
		return fmt.Errorf("current and new name are the same")
	}

	svcs, sess, err := cmdDeps.RequireAuth()
	if err != nil {
		return err
	}
	trays, err := svcs.Trays.ListMine(cmd.Context(), sess)
	if err != nil {
		return err
	}
	configDir := cmdDeps.ConfigDir()
	aliasesMap := cmdDeps.RemoteAliases()

	f, err := remotesfile.Load(configDir)
	if err != nil {
		return err
	}
	if f.Aliases == nil {
		f.Aliases = map[string]string{}
	}

	var foundOldKey string
	var trayID string
	for k, v := range f.Aliases {
		if strings.EqualFold(k, curr) {
			foundOldKey = k
			trayID = strings.TrimSpace(v)
			break
		}
	}

	if foundOldKey != "" {
		for k, v := range f.Aliases {
			if !strings.EqualFold(k, newName) {
				continue
			}
			if strings.TrimSpace(v) != trayID {
				return fmt.Errorf("local alias %q already points to a different tray", newName)
			}
			// newName already maps to this tray — drop the old key only.
			delete(f.Aliases, foundOldKey)
			if err := remotesfile.Save(configDir, f); err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Removed alias %q; %q already refers to this tray.\n", foundOldKey, k)
			return err
		}
		delete(f.Aliases, foundOldKey)
		f.Aliases[newName] = trayID
		if err := remotesfile.Save(configDir, f); err != nil {
			return err
		}
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Renamed remote alias %q → %q (tray %s).\n", foundOldKey, newName, trayID)
		return err
	}

	tid, err := trayref.ResolveTrayRef(cmd.Context(), svcs, sess, curr, aliasesMap)
	if err != nil {
		return fmt.Errorf("no remote alias %q and could not resolve as a tray — run `tray ls` or `tray remote ls`", curr)
	}
	if _, ok := trayref.TrayByID(trays, tid); !ok {
		return fmt.Errorf("tray not in your list — run `tray ls` (you must already be a member or owner)")
	}
	for k, v := range f.Aliases {
		if strings.EqualFold(k, newName) && strings.TrimSpace(v) != tid {
			return fmt.Errorf("local alias %q already points to a different tray", newName)
		}
	}
	f.Aliases[newName] = tid
	if err := remotesfile.Save(configDir, f); err != nil {
		return err
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Saved local alias %q → tray %s (you can use it anywhere a tray name goes).\n", newName, tid)
	return err
}

func runRemoteLs(cmd *cobra.Command, args []string) error {
	f, err := remotesfile.Load(cmdDeps.ConfigDir())
	if err != nil {
		return err
	}
	format, err := output.FormatFromCmd(cmd)
	if err != nil {
		return err
	}
	switch format {
	case output.FormatJSON:
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(f.Aliases)
	case output.FormatMarkdown:
		if len(f.Aliases) == 0 {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), "_No remotes._")
			return err
		}
		_, err := fmt.Fprintf(cmd.OutOrStdout(), "| %s | %s |\n", "Alias", "Tray ID")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "| %s | %s |\n", "---", "---")
		if err != nil {
			return err
		}
		keys := sortedKeys(f.Aliases)
		for _, k := range keys {
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "| %s | `%s` |\n",
				strings.ReplaceAll(k, "|", "\\|"), f.Aliases[k])
			if err != nil {
				return err
			}
		}
		return nil
	default:
		if len(f.Aliases) == 0 {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), "No remotes.")
			return err
		}
		tw := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		_, err := fmt.Fprintln(tw, "ALIAS\tTRAY_ID")
		if err != nil {
			return err
		}
		for _, k := range sortedKeys(f.Aliases) {
			_, err := fmt.Fprintf(tw, "%s\t%s\n", k, f.Aliases[k])
			if err != nil {
				return err
			}
		}
		return tw.Flush()
	}
}

func runRemoteRemove(cmd *cobra.Command, args []string) error {
	key := strings.ToLower(strings.TrimSpace(args[0]))
	if key == "" {
		return fmt.Errorf("which alias should we remove?")
	}
	f, err := remotesfile.Load(cmdDeps.ConfigDir())
	if err != nil {
		return err
	}
	var found string
	for k := range f.Aliases {
		if strings.ToLower(k) == key {
			found = k
			break
		}
	}
	if found == "" {
		return fmt.Errorf("no remote alias matches %q", args[0])
	}
	delete(f.Aliases, found)
	if err := remotesfile.Save(cmdDeps.ConfigDir(), f); err != nil {
		return err
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Removed remote %q.\n", found)
	return err
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
