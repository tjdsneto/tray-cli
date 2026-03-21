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
var loginVerbose bool

func cmdLogin() *cobra.Command {
	c := &cobra.Command{
		Use:   "login",
		Short: "Sign in with OAuth (browser) or a pasted access token",
		Long: `Signs in to Supabase Auth and stores credentials under the tray config directory.

Requires ` + config.EnvSupabaseURL + ` and ` + config.EnvSupabaseAnonKey + ` (environment) or embeds from ./run.sh and ./build.sh with a repo-root .env.

**OAuth:** run without --provider to open a **local web page** where you pick Google, GitHub, etc. Or pass --provider (or ` + config.EnvOAuthProvider + `) to skip the picker and use one provider. Enable each provider under Supabase → Authentication → Providers.

Supabase (final redirect to this CLI): under Authentication → URL Configuration → Redirect URLs, allow local callbacks, e.g. http://127.0.0.1:*/** or the exact "Listening for callback" URL printed at runtime.

In Google Cloud / GitHub OAuth settings (etc.): authorized redirect URI must be Supabase’s callback only — no wildcards, no localhost — e.g.:
  https://<your-project-ref>.supabase.co/auth/v1/callback
The provider redirects there; Supabase then redirects your browser to localhost with the auth code.

**Manual:** use --token with a user JWT from the Supabase dashboard, curl password grant, or another client.`,
		Example: `  tray login
  tray login --provider google
  TRAY_OAUTH_PROVIDER=github ./run.sh login
  tray login --token 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...'`,
		RunE: runLogin,
	}
	c.Flags().StringVar(&loginToken, "token", "", "skip OAuth and use this user access token (JWT)")
	c.Flags().StringVar(&loginProvider, "provider", "", "skip the provider picker and sign in with this id (e.g. google); optional if you use "+config.EnvOAuthProvider+" or the web picker")
	c.Flags().BoolVarP(&loginVerbose, "verbose", "v", false, "print detailed one-time Supabase / Google Cloud OAuth setup hints")
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

func effectiveOAuthProvider() string {
	if s := strings.TrimSpace(loginProvider); s != "" {
		return s
	}
	return config.OAuthProvider()
}

func runLoginOAuth(cmd *cobra.Command, url, key string) error {
	provider := effectiveOAuthProvider()
	if loginVerbose {
		if provider != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "OAuth sign-in (provider=%s).\n", provider)
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "OAuth sign-in — pick a provider in the browser.\n")
		}
		fmt.Fprintf(cmd.OutOrStdout(), "One-time setup:\n")
		fmt.Fprintf(cmd.OutOrStdout(), "  • Supabase → Authentication → URL Configuration → Redirect URLs: allow local callbacks, e.g. http://127.0.0.1:*/**\n")
		fmt.Fprintf(cmd.OutOrStdout(), "  • Google/GitHub OAuth app → Authorized redirect URI (use Supabase’s URL, not localhost):\n    %s\n\n", supabaseAuthCallbackURL(url))
	} else {
		if provider != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "Signing in with %s…\n", provider)
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "Opening sign-in in your browser…\n")
		}
	}

	accessToken, refreshToken, userID, email, err := auth.LoginWithOAuth(
		cmd.Context(),
		url,
		key,
		provider,
		nil,
		func(callbackURL string, pickURL string, direct string) {
			if !loginVerbose {
				return
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Local callback (for this session only): %s\n", callbackURL)
			if direct != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Opening provider sign-in…\n")
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "Opening provider picker: %s\n", pickURL)
			}
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
