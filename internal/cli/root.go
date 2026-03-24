package cli

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/tjdsneto/tray-cli/internal/cli/commands"
	"github.com/tjdsneto/tray-cli/internal/output"
	"github.com/tjdsneto/tray-cli/internal/remotesfile"
)

// Execute runs the tray root command. Errors are formatted in cmd/tray/main.go (user-facing line + optional debug).
func Execute() error {
	return NewRootCmd().ExecuteContext(context.Background())
}

// NewRootCmd builds the full CLI tree.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "tray",
		Short: "Shared task inbox — capture, triage, and share work",
		Long: `Tray helps you manage a shared task inbox with named trays.

Run "tray help" for commands, or "tray <command> --help" for options.`,
		// SilenceUsage: when a RunE returns an error, do not print the full command usage block (keeps stderr short).
		SilenceUsage: true,
		// SilenceErrors: cobra does not print the raw error; cmd/tray prints a single user-facing line (and debug detail when TRAY_DEBUG=1).
		SilenceErrors: true,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	// --config-dir: optional override; default path follows XDG on Unix and APPDATA on Windows (see internal/config/paths if exposed).
	root.PersistentFlags().StringVar(&configDirFlag, "config-dir", "", "override where tray stores local files (credentials and remotes)")

	// Registers --format, deprecated -o/--output, and --json on the root (inherited by subcommands).
	output.RegisterPersistentFlags(root.PersistentFlags())

	commands.Register(root, commands.Deps{
		RequireAuth: requireAuth,
		ConfigDir:   ConfigDir,
		RemoteAliases: func() map[string]string {
			return remotesfile.AliasesMap(ConfigDir())
		},
	})

	return root
}

var configDirFlag string
