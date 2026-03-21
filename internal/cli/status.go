package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tjdsneto/tray-cli/internal/auth"
	"github.com/tjdsneto/tray-cli/internal/config"
	"github.com/tjdsneto/tray-cli/internal/credentials"
	"github.com/tjdsneto/tray-cli/internal/output"
)

func cmdStatus() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show whether you are signed in",
		Long:  `Checks for saved credentials and validates the session with Supabase. Exit code 0 if signed in, 1 if not.`,
		RunE:  runStatus,
	}
}

type statusJSON struct {
	SignedIn bool   `json:"signed_in"`
	Email    string `json:"email,omitempty"`
}

func runStatus(cmd *cobra.Command, _ []string) error {
	url := config.SupabaseURL()
	key := config.SupabaseAnonKey()
	if url == "" || key == "" {
		return fmt.Errorf("set %s and %s (environment) or build with ./run.sh or ./build.sh and a .env", config.EnvSupabaseURL, config.EnvSupabaseAnonKey)
	}
	format, err := output.FormatFromCmd(cmd)
	if err != nil {
		return err
	}

	cred, err := credentials.Load(ConfigDir())
	if err != nil {
		return writeStatusFailure(cmd, format, err)
	}
	user, err := auth.FetchUser(cmd.Context(), url, key, cred.AccessToken, nil)
	if err != nil {
		return writeStatusFailure(cmd, format, err)
	}

	switch format {
	case output.FormatJSON:
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(statusJSON{SignedIn: true, Email: strings.TrimSpace(user.Email)})
	case output.FormatMarkdown:
		_, err := fmt.Fprintf(cmd.OutOrStdout(), "| | |\n|--|--|\n| signed_in | yes |\n| email | %s |\n", mdEscape(strings.TrimSpace(user.Email)))
		return err
	default:
		if e := strings.TrimSpace(user.Email); e != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "Signed in as %s.\n", e)
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "Signed in.")
		}
		return nil
	}
}

func mdEscape(s string) string {
	if s == "" {
		return "—"
	}
	return strings.ReplaceAll(strings.ReplaceAll(s, "|", "\\|"), "\n", " ")
}

func writeStatusFailure(cmd *cobra.Command, format output.Format, cause error) error {
	_ = cause // do not leak API bodies to stdout
	switch format {
	case output.FormatJSON:
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		_ = enc.Encode(statusJSON{SignedIn: false})
	case output.FormatMarkdown:
		fmt.Fprintln(cmd.OutOrStdout(), "| | |\n|--|--|\n| signed_in | no |")
	default:
		// stderr: main prints "tray: not signed in"
	}
	return fmt.Errorf("not signed in")
}
