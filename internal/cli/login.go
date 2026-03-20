package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tjdsneto/tray-cli/internal/auth"
	"github.com/tjdsneto/tray-cli/internal/config"
	"github.com/tjdsneto/tray-cli/internal/credentials"
)

var loginToken string

func cmdLogin() *cobra.Command {
	c := &cobra.Command{
		Use:   "login",
		Short: "Save a Supabase session using an access token",
		Long: `Validates the token against Supabase Auth and stores credentials under the tray config directory.

Requires Supabase settings via ` + config.EnvSupabaseURL + ` and ` + config.EnvSupabaseAnonKey + ` (environment), or embed them at build time (see ./run.sh and ./build.sh with a repo-root .env).

Use --token with a JWT from your app (e.g. after OAuth) or from the Supabase dashboard while testing.
Browser-based login can be added later.`,
		RunE: runLogin,
	}
	c.Flags().StringVar(&loginToken, "token", "", "Supabase access token (JWT)")
	_ = c.MarkFlagRequired("token")
	return c
}

func runLogin(_ *cobra.Command, _ []string) error {
	url := config.SupabaseURL()
	key := config.SupabaseAnonKey()
	if url == "" || key == "" {
		return fmt.Errorf("set %s and %s (environment) or build with ./run.sh or ./build.sh and a .env", config.EnvSupabaseURL, config.EnvSupabaseAnonKey)
	}
	user, err := auth.FetchUserTimeout(url, key, loginToken, nil)
	if err != nil {
		return err
	}
	dir := ConfigDir()
	if err := credentials.Save(dir, credentials.File{
		AccessToken: loginToken,
		UserID:      user.ID,
	}); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "Logged in as %s (%s)\n", user.Email, user.ID)
	return nil
}
