package commands

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

const installScriptURL = "https://raw.githubusercontent.com/tjdsneto/tray-cli/main/scripts/install.sh"

var runUpgradeScript = defaultRunUpgradeScript

func cmdUpgrade() *cobra.Command {
	var version string
	c := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade tray using the install script",
		Long: `Re-runs Tray's install script to update your local binary.

This command currently supports only the install-script distribution method.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpgrade(cmd, strings.TrimSpace(version))
		},
	}
	c.Flags().StringVar(&version, "version", "latest", "tray version to install (for example: v0.1.1)")
	return c
}

func runUpgrade(cmd *cobra.Command, version string) error {
	if runtime.GOOS == "windows" {
		return fmt.Errorf("`tray upgrade` currently supports install-script setups on Unix-like systems only")
	}
	if version == "" {
		version = "latest"
	}
	if strings.ContainsAny(version, " \t\r\n") {
		return fmt.Errorf("invalid --version value %q", version)
	}
	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Upgrading tray via install script (TRAY_VERSION=%s)...\n", version); err != nil {
		return err
	}
	if err := runUpgradeScript(cmd, version); err != nil {
		return err
	}
	_, err := fmt.Fprintln(cmd.OutOrStdout(), "Upgrade completed. Run `tray --version` to verify.")
	return err
}

func defaultRunUpgradeScript(cmd *cobra.Command, version string) error {
	command := "curl -fsSL " + installScriptURL + " | bash"
	ec := exec.CommandContext(cmd.Context(), "sh", "-c", command)
	ec.Stdout = cmd.OutOrStdout()
	ec.Stderr = cmd.ErrOrStderr()
	ec.Stdin = os.Stdin
	env := os.Environ()
	env = append(env, "TRAY_VERSION="+version)
	ec.Env = env
	if err := ec.Run(); err != nil {
		return fmt.Errorf("upgrade failed: %w", err)
	}
	return nil
}
