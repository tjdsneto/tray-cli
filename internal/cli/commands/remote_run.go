package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/tjdsneto/tray-cli/internal/cli/trayref"
	"github.com/tjdsneto/tray-cli/internal/domain"
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
		return fmt.Errorf("tray not in your list — run `tray remote ls` (joined) or `tray ls` (owned)")
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
	svcs, sess, err := cmdDeps.RequireAuth()
	if err != nil {
		return err
	}
	joined, err := svcs.Trays.ListJoined(cmd.Context(), sess)
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
	byTray := aliasesByTrayID(f.Aliases)
	joinedIDs := make(map[string]struct{}, len(joined))
	for i := range joined {
		joinedIDs[strings.TrimSpace(joined[i].ID)] = struct{}{}
	}
	orphans := map[string]string{}
	for alias, tid := range f.Aliases {
		tid = strings.TrimSpace(tid)
		if _, ok := joinedIDs[tid]; !ok {
			orphans[alias] = tid
		}
	}

	format, err := output.FormatFromCmd(cmd)
	if err != nil {
		return err
	}
	switch format {
	case output.FormatJSON:
		payload := buildRemoteListPayload(joined, byTray, orphans)
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(payload)
	case output.FormatMarkdown:
		return writeRemoteLsMarkdown(cmd.OutOrStdout(), joined, byTray, orphans)
	default:
		return writeRemoteLsTable(cmd.OutOrStdout(), joined, byTray, orphans)
	}
}

type remoteListPayload struct {
	Joined        []remoteJoinedPayload `json:"joined"`
	OrphanAliases map[string]string     `json:"orphan_aliases,omitempty"`
}

type remoteJoinedPayload struct {
	TrayID   string   `json:"tray_id"`
	Name     string   `json:"name"`
	OwnerID  string   `json:"owner_id"`
	Aliases  []string `json:"aliases"`
	JoinedAt *string  `json:"joined_at,omitempty"`
}

func buildRemoteListPayload(joined []domain.Tray, byTray map[string][]string, orphans map[string]string) remoteListPayload {
	out := remoteListPayload{Joined: make([]remoteJoinedPayload, 0, len(joined))}
	for i := range joined {
		t := &joined[i]
		aliases := byTray[strings.TrimSpace(t.ID)]
		if aliases == nil {
			aliases = []string{}
		}
		rp := remoteJoinedPayload{
			TrayID:  strings.TrimSpace(t.ID),
			Name:    t.Name,
			OwnerID: strings.TrimSpace(t.OwnerID),
			Aliases: aliases,
		}
		if t.MemberJoinedAt != nil {
			s := t.MemberJoinedAt.UTC().Format(time.RFC3339)
			rp.JoinedAt = &s
		}
		out.Joined = append(out.Joined, rp)
	}
	if len(orphans) > 0 {
		out.OrphanAliases = orphans
	}
	return out
}

func aliasesByTrayID(aliases map[string]string) map[string][]string {
	by := make(map[string][]string)
	for a, tid := range aliases {
		tid = strings.TrimSpace(tid)
		by[tid] = append(by[tid], a)
	}
	for k := range by {
		sort.Strings(by[k])
	}
	return by
}

func writeRemoteLsMarkdown(w io.Writer, joined []domain.Tray, byTray map[string][]string, orphans map[string]string) error {
	if len(joined) == 0 {
		_, err := fmt.Fprintln(w, "_No joined trays._")
		if err != nil {
			return err
		}
	} else {
		_, err := fmt.Fprintf(w, "| %s | %s | %s | %s |\n", "Name", "Aliases", "Tray ID", "Joined")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(w, "| %s | %s | %s | %s |\n", "---", "---", "---", "---")
		if err != nil {
			return err
		}
		for i := range joined {
			t := &joined[i]
			al := strings.Join(byTray[strings.TrimSpace(t.ID)], ", ")
			if al == "" {
				al = "—"
			}
			jd := "—"
			if t.MemberJoinedAt != nil {
				jd = output.FormatTrayLocalTime(*t.MemberJoinedAt)
			}
			_, err := fmt.Fprintf(w, "| %s | %s | `%s` | %s |\n",
				strings.ReplaceAll(t.Name, "|", "\\|"),
				strings.ReplaceAll(al, "|", "\\|"),
				strings.TrimSpace(t.ID),
				jd,
			)
			if err != nil {
				return err
			}
		}
	}
	if len(orphans) > 0 {
		_, err := fmt.Fprint(w, "\n_Orphan local aliases (tray not joined or left):_\n")
		if err != nil {
			return err
		}
		for _, k := range sortedKeys(orphans) {
			_, err := fmt.Fprintf(w, "- `%s` → `%s`\n", k, orphans[k])
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func writeRemoteLsTable(w io.Writer, joined []domain.Tray, byTray map[string][]string, orphans map[string]string) error {
	if len(joined) == 0 && len(orphans) == 0 {
		_, err := fmt.Fprint(w, `No joined trays.

Join with an invite:  tray join <token-or-url> [local-alias]
Optional nickname:    tray remote add <alias> <token-or-url>
`)
		return err
	}
	if len(joined) > 0 {
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		_, err := fmt.Fprintln(tw, "NAME\tALIASES\tTRAY_ID\tJOINED")
		if err != nil {
			return err
		}
		for i := range joined {
			t := &joined[i]
			al := strings.Join(byTray[strings.TrimSpace(t.ID)], ", ")
			if al == "" {
				al = "—"
			}
			jd := "—"
			if t.MemberJoinedAt != nil {
				jd = output.FormatTrayLocalTime(*t.MemberJoinedAt)
			}
			_, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", t.Name, al, strings.TrimSpace(t.ID), jd)
			if err != nil {
				return err
			}
		}
		if err := tw.Flush(); err != nil {
			return err
		}
	}
	if len(orphans) > 0 {
		_, err := fmt.Fprintln(w, "\nOrphan local aliases (tray not joined or left):")
		if err != nil {
			return err
		}
		for _, k := range sortedKeys(orphans) {
			_, err := fmt.Fprintf(w, "  %s  →  %s\n", k, orphans[k])
			if err != nil {
				return err
			}
		}
	}
	return nil
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
