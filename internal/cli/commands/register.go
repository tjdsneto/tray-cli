package commands

import (
	"github.com/spf13/cobra"
)

// Register attaches all subcommands to root. Call once during startup; deps must outlive Execute.
func Register(root *cobra.Command, deps Deps) {
	cmdDeps = deps

	// --- Session & account ---
	root.AddCommand(cmdLogin(), cmdStatus())
	root.AddCommand(cmdUpgrade())

	// --- Trays: lifecycle & sharing ---
	root.AddCommand(cmdCreate(), cmdLs(), cmdRename(), cmdDeleteTray())
	root.AddCommand(cmdInvite(), cmdRotateInvite(), cmdJoin())

	// --- Items ---
	root.AddCommand(cmdAdd(), cmdList(), cmdRemove(), cmdContributed())

	// --- Remotes & membership ---
	root.AddCommand(cmdRemote())
	root.AddCommand(cmdMembers(), cmdRevoke(), cmdLeave())

	// --- Triage & stubs ---
	root.AddCommand(cmdReview(), cmdTriage())
	root.AddCommand(cmdAccept(), cmdDecline())
	root.AddCommand(cmdSnooze(), cmdComplete(), cmdArchive())
	root.AddCommand(cmdListen())
}
