package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tjdsneto/tray-cli/internal/auth"
	"github.com/tjdsneto/tray-cli/internal/config"
	"github.com/tjdsneto/tray-cli/internal/credentials"
)

var loginToken string
var loginProvider string

func cmdLogin() *cobra.Command {
	c := &cobra.Command{
		Use:   "login",
		Short: "Sign in with OAuth (browser) or a pasted access token",
		Long: `Signs in to Supabase Auth and stores credentials under the tray config directory.

Requires ` + config.EnvSupabaseURL + ` and ` + config.EnvSupabaseAnonKey + ` (environment) or embeds from ./run.sh and ./build.sh with a repo-root .env.

**Default:** opens your browser for OAuth (PKCE). Enable the provider under Supabase → Authentication → Providers.

Supabase (final redirect to this CLI): under Authentication → URL Configuration → Redirect URLs, allow local callbacks, e.g. http://127.0.0.1:*/** or the exact "Listening for callback" URL printed at runtime.

In Google Cloud / GitHub OAuth settings (etc.): authorized redirect URI must be Supabase’s callback only — no wildcards, no localhost — e.g.:
  https://<your-project-ref>.supabase.co/auth/v1/callback
The provider redirects there; Supabase then redirects your browser to localhost with the auth code.

**Manual:** use --token with a user JWT from the Supabase dashboard, curl password grant, or another client.`,
		Example: `  tray login
  tray login --provider google
  tray login --token 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...'`,
		RunE: runLogin,
	}
	c.Flags().StringVar(&loginToken, "token", "", "skip OAuth and use this user access token (JWT)")
	c.Flags().StringVar(&loginProvider, "provider", "github", "OAuth provider id (e.g. github, google) — must be enabled in Supabase")
	return c
}

func runLogin(cmd *cobra.Command, _ []string) error {
	url := config.SupabaseURL()
	key := config.SupabaseAnonKey()
	if url == "" || key == "" {
		return fmt.Errorf("set %s and %s (environment) or build with ./run.sh or ./build.sh and a .env", config.EnvSupabaseURL, config.EnvSupabaseAnonKey)
	}
	if strings.TrimSpace(loginToken) != "" {
		return runLoginWithToken(url, key)
	}
	return runLoginOAuth(cmd, url, key)
}

func runLoginOAuth(cmd *cobra.Command, url, key string) error {
	fmt.Fprintf(cmd.OutOrStdout(), "OAuth sign-in (provider=%s).\n", loginProvider)
	fmt.Fprintf(cmd.OutOrStdout(), "Supabase → Authentication → URL Configuration → Redirect URLs (allow the CLI callback), e.g.:\n  http://127.0.0.1:*/**\n")
	fmt.Fprintf(cmd.OutOrStdout(), "Google/GitHub/etc. OAuth app: authorized redirect must be Supabase’s callback (no wildcards, not localhost):\n  %s\n", supabaseAuthCallbackURL(url))
	fmt.Fprintln(cmd.OutOrStdout())

	accessToken, refreshToken, userID, email, err := auth.LoginWithOAuth(
		cmd.Context(),
		url,
		key,
		strings.TrimSpace(loginProvider),
		nil,
		func(callbackURL string) {
			fmt.Fprintf(cmd.OutOrStdout(), "Listening for callback: %s\n", callbackURL)
			fmt.Fprintf(cmd.OutOrStdout(), "Opening browser…\n")
		},
	)
	if err != nil {
		return err
	}
	dir := ConfigDir()
	if err := credentials.Save(dir, credentials.File{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserID:       userID,
	}); err != nil {
		return err
	}
	label := strings.TrimSpace(email)
	if label == "" {
		label = userID
	}
	fmt.Fprintf(os.Stdout, "Logged in as %s (%s)\n", label, userID)
	return nil
}

func runLoginWithToken(url, key string) error {
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
	label := strings.TrimSpace(user.Email)
	if label == "" {
		label = user.ID
	}
	fmt.Fprintf(os.Stdout, "Logged in as %s (%s)\n", label, user.ID)
	return nil
}

func supabaseAuthCallbackURL(projectURL string) string {
	return strings.TrimRight(strings.TrimSpace(projectURL), "/") + "/auth/v1/callback"
}
