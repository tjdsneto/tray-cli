package cli

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/tjdsneto/tray-cli/internal/output"
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
	svcs, sess, err := requireAuth()
	if err != nil {
		return err
	}
	trayID, err := svcs.Trays.Join(cmd.Context(), sess, tok)
	if err != nil {
		return err
	}
	f, err := loadRemotes(ConfigDir())
	if err != nil {
		return err
	}
	if f.Aliases == nil {
		f.Aliases = map[string]string{}
	}
	f.Aliases[alias] = trayID
	if err := saveRemotes(ConfigDir(), f); err != nil {
		return err
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Saved alias %q → tray %s (you can use the alias anywhere a tray name goes).\n", alias, trayID)
	return err
}

func runRemoteLs(cmd *cobra.Command, args []string) error {
	f, err := loadRemotes(ConfigDir())
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
	f, err := loadRemotes(ConfigDir())
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
	if err := saveRemotes(ConfigDir(), f); err != nil {
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
