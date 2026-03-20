package cli

import "github.com/spf13/cobra"

func cmdLogin() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Authenticate with Supabase (browser or device flow)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return stub("login")
		},
	}
}
