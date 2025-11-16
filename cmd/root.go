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
	"os/exec"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"

	"github.com/commoddity/devint/cmd/config"
	"github.com/commoddity/devint/cmd/git"
)

// getVersion returns the version string from build info, git, or falls back to "dev"
func getVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		// Fallback to git describe for local builds
		return getVersionFromGit()
	}

	const modulePath = "github.com/commoddity/devint"

	// First, check if this is the main module (when built from source)
	if info.Main.Path == modulePath {
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			version := strings.TrimPrefix(info.Main.Version, "v")
			if version != "" {
				return version
			}
		}
	}

	// For installed binaries, check dependencies to find our module version
	// This handles cases like `go install github.com/commoddity/devint@latest`
	// where the module appears as a dependency in the build info
	for _, dep := range info.Deps {
		if dep.Path == modulePath {
			if dep.Version != "" && dep.Version != "(devel)" {
				version := strings.TrimPrefix(dep.Version, "v")
				if version != "" {
					return version
				}
			}
		}
	}

	// Fallback: try git describe for local development builds
	return getVersionFromGit()
}

// getVersionFromGit attempts to get version from git describe
func getVersionFromGit() string {
	cmd := exec.Command("git", "describe", "--tags", "--always", "--dirty")
	output, err := cmd.Output()
	if err != nil {
		return "dev"
	}
	version := strings.TrimSpace(string(output))
	version = strings.TrimPrefix(version, "v")
	version = strings.TrimSpace(version)
	if version == "" {
		return "dev"
	}
	return version
}

var rootCmd = &cobra.Command{
	Use:   "devint",
	Short: "Developer Interface - streamline your development workflows",
	Long: `Developer Interface (DI) is a comprehensive CLI tool designed for development
	workflows. It provides users with a unified approach to manage configuration
	settings, execute Git operations, and interact with integrated LLM providers for tasks
	such as automated pull request generation.`,
	Version: getVersion(),
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Toggle verbose mode or other options")
	rootCmd.AddCommand(git.GitCmd)
	rootCmd.AddCommand(config.ConfigCmd)
	rootCmd.AddCommand(chatCmd)

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
