package git

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/go-git/go-git/v5"
)

// Command templates and constants used to generate a unified diff.
// These commands are combined and sent to an LLM to provide context for the diff generation.
const (
	// gitDiffCmdTemplate generates a diff for the repository using the given repository root and target branch.
	gitDiffCmdTemplate = `git -C %s --no-pager diff %s --unified=0 -- .`
	// grepCmd filters out metadata lines from the diff output.
	grepCmd = `grep -vE '^(diff --git|index |@@)'`
	// sedCmd reformats file header lines in the diff for better readability.
	sedCmd = `sed -E 's/^--- a\//Old File: /; s/^\+\+\+ b\//New File: /'`
	// finalGrepCmd removes any remaining empty lines from the diff output.
	finalGrepCmd = `grep -vE '^$'`

	// CombinedDiffCmd aggregates the above commands into one string.
	// This variable is provided to the LLM for context.
	CombinedDiffCmd = gitDiffCmdTemplate + " | " + grepCmd + " | " + sedCmd + " | " + finalGrepCmd
)

// GenerateDiff creates a unified diff between the current branch and the target branch.
// It executes multiple shell commands to generate, filter, and format the diff output.
// The final diff is wrapped in a markdown diff code block.
func (p *Provider) GenerateDiff(ctx context.Context, targetBranch string) (string, error) {
	// Obtain the repository root directory.
	repoRoot, err := p.getRepoRoot()
	if err != nil {
		return "", fmt.Errorf("failed to get repository root: %w", err)
	}

	// Build the git diff command.
	gitDiffCmd := fmt.Sprintf(gitDiffCmdTemplate, repoRoot, targetBranch)
	p.logger.Info("🔍 Executing git diff command...", "target_branch", targetBranch)
	gitDiffOutput, err := exec.Command("bash", "-c", gitDiffCmd).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to execute git diff command: %v\nOutput: %s", err, string(gitDiffOutput))
	}

	// Filter the diff output to remove unwanted metadata lines using grep.
	grepCmdProc := exec.Command("bash", "-c", grepCmd)
	grepCmdProc.Stdin = bytes.NewReader(gitDiffOutput)
	grepOutput, err := grepCmdProc.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to execute grep command: %v\nOutput: %s", err, string(grepOutput))
	}

	// Reformat the output using sed for better clarity.
	sedCmdProc := exec.Command("bash", "-c", sedCmd)
	sedCmdProc.Stdin = bytes.NewReader(grepOutput)
	sedOutput, err := sedCmdProc.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to execute sed command: %v\nOutput: %s", err, string(sedOutput))
	}

	// Remove any empty lines to minimize noise in the output.
	finalGrepCmdProc := exec.Command("bash", "-c", finalGrepCmd)
	finalGrepCmdProc.Stdin = bytes.NewReader(sedOutput)
	finalOutput, err := finalGrepCmdProc.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to execute final grep command: %v\nOutput: %s", err, string(finalOutput))
	}

	// Sanitize the diff to remove identifying information before sending to LLM.
	sanitizedOutput := sanitizeDiff(string(finalOutput), p.companyName)

	// Wrap the final output in a markdown diff code block.
	wrappedOutput := fmt.Sprintf("```diff\n%s\n```", sanitizedOutput)
	return wrappedOutput, nil
}

// GetPRDiff retrieves the diff for a pull request by number using the GitHub CLI.
// It executes `gh pr diff` command and formats the output similar to GenerateDiff.
func (p *Provider) GetPRDiff(ctx context.Context, number int) (string, error) {
	if number == 0 {
		return "", fmt.Errorf("pull request number is required")
	}

	// Use gh CLI to get the PR diff
	p.logger.Info("🔍 Fetching PR diff...", "pr_number", number)
	cmd := exec.Command("gh", "pr", "diff", fmt.Sprintf("%d", number))
	diffOutput, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get PR diff: %w\nOutput: %s", err, string(diffOutput))
	}

	// Filter the diff output to remove unwanted metadata lines using grep.
	grepCmdProc := exec.Command("bash", "-c", grepCmd)
	grepCmdProc.Stdin = bytes.NewReader(diffOutput)
	grepOutput, err := grepCmdProc.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to execute grep command: %w\nOutput: %s", err, string(grepOutput))
	}

	// Reformat the output using sed for better clarity.
	sedCmdProc := exec.Command("bash", "-c", sedCmd)
	sedCmdProc.Stdin = bytes.NewReader(grepOutput)
	sedOutput, err := sedCmdProc.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to execute sed command: %w\nOutput: %s", err, string(sedOutput))
	}

	// Remove any empty lines to minimize noise in the output.
	finalGrepCmdProc := exec.Command("bash", "-c", finalGrepCmd)
	finalGrepCmdProc.Stdin = bytes.NewReader(sedOutput)
	finalOutput, err := finalGrepCmdProc.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to execute final grep command: %w\nOutput: %s", err, string(finalOutput))
	}

	// Sanitize the diff to remove identifying information before sending to LLM.
	sanitizedOutput := sanitizeDiff(string(finalOutput), p.companyName)

	// Wrap the final output in a markdown diff code block.
	wrappedOutput := fmt.Sprintf("```diff\n%s\n```", sanitizedOutput)
	return wrappedOutput, nil
}

// getRepoRoot returns the absolute path of the repository root.
// It uses the go-git library to open the repository and locate the worktree.
func (p *Provider) getRepoRoot() (string, error) {
	// Get the current working directory.
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Open the Git repository based on the current directory.
	repo, err := git.PlainOpen(dir)
	if err != nil {
		return "", err
	}

	// Access the repository's worktree.
	worktree, err := repo.Worktree()
	if err != nil {
		return "", err
	}

	// Return the root directory of the repository.
	repoRoot := worktree.Filesystem.Root()
	return repoRoot, nil
}
