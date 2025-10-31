// ---------------------------------------------------------------------------
// File: root.go
// Package: cmd
//
// Overview:
//
// The Developer Interface (DI) is a command-line tool designed to streamline
// developer workflows. DI is intended to help developers quickly perform
// routine operations and maintain consistency across projects.
//
// DEV_NOTE:
//
// This repo is also intended to be a living project and should be updated to incorporate
// any scripts, hacks, time-saving features, etc. that are useful in local
// development workflows and that could benefit others using this CLI
//
// ---------------------------------------------------------------------------
package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/commoddity/devint/cmd/config"
	"github.com/commoddity/devint/cmd/git"
)

var rootCmd = &cobra.Command{
	Use:   "devint",
	Short: "Developer Interface - streamline your development workflows",
	Long: `Developer Interface (DI) is a comprehensive CLI tool designed for development
	workflows. It provides users with a unified approach to manage configuration
	settings, execute Git operations, and interact with integrated LLM providers for tasks
	such as automated pull request generation.`,
	Version: "0.0.1",
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Toggle verbose mode or other options")
	rootCmd.AddCommand(git.GitCmd)
	rootCmd.AddCommand(config.ConfigCmd)

	if !config.ConfigExists() {
		config.RunFirstTimeSetup()
		return
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
