// ---------------------------------------------------------------------------
// File: root.go
// Package: git
//
// Purpose:
//   This file defines the main "git" command used by the Developer Interface.
//   It acts as a container for various git-related subcommands (such as creating a PR)
//   that help users interact with git. This command is built using Cobra, and it
//   automatically displays help information if no subcommand is provided.
//
// Features:
//   - Provides a brief usage overview for git-related operations.
//   - Delegates functionality to subcommands (e.g., createpr).
//
// Usage:
//   Running the command "devint git" will display help information if no subcommand
//   is specified. Subcommands can be used to generate pull requests and to perform
//   other git operations.
// ---------------------------------------------------------------------------

package git

import (
	"github.com/spf13/cobra"
)

// GitCmd represents the main git command.
// It acts as a parent for all git-related subcommands.
var GitCmd = &cobra.Command{
	Use:   "git",
	Short: "A collection of git commands to help you work with git.",
	Long: `A collection of git commands to help you work with git.

This command aggregates several subcommands that allow you to interact with git,
such as creating a pull request, displaying git diffs, and managing branches.
If no subcommand is provided, the command will display its help information.`,
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommand is specified, display help.
		if len(args) == 0 {
			cmd.Help()
			return
		}
	},
}

// The init function adds all available subcommands to the GitCmd.
func init() {
	// Add the createpr subcommand to the main git command.
	GitCmd.AddCommand(createprCmd)
}
