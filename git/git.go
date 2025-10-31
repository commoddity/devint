// ---------------------------------------------------------------------------
// File: git.go
// Package: git
//
// Purpose:
//
//	This file implements key functionalities for interacting with Git repositories
//	and GitHub. It provides a Provider struct that encapsulates the GitHub client
//	and logger, as well as functions to create pull requests, push branches to remote,
//	and generate formatted diffs for use with a language model (LLM).
//
// Features:
//   - Validates Git configuration and initializes a Git provider with authentication.
//   - Creates GitHub pull requests after ensuring the current branch is pushed.
//   - Generates unified diffs between branches with several shell commands to format
//     the output.
//   - Offers utility functions for obtaining repository metadata like root directory,
//     current branch name, and repository name using the go-git library.
//
// ---------------------------------------------------------------------------
package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/google/go-github/v69/github"
	"github.com/pokt-network/poktroll/pkg/polylog"

	gitCfg "github.com/commoddity/devint/config/git"
)

// repoOwner is set per-request through Provider methods that accept it as a parameter.

var (
	// Suggest configuring a valid Personal Access Token for GitHub if attempting to perform operations on a private repository.
	suggestConfiguringPAT = "If the failure is due to a missing or invalid Personal Access Token, configure a valid Personal Access Token for GitHub in your config file.\nYou may do so by running `devint config`."

	errPullRequestFailed = errors.New("git error: pull request failed")
)

// Provider represents a Git provider that encapsulates a GitHub client
// and a logger. It provides methods to create pull requests and to
// interact with Git repositories.
type Provider struct {
	logger         polylog.Logger // Logger for logging operations.
	*github.Client                // Embedded GitHub client for API calls.
}

// NewGitProvider initializes and returns a new Git provider.
// It validates the provided Git configuration and sets up an authenticated GitHub client.
func NewGitProvider(logger polylog.Logger, cfg gitCfg.Config) (*Provider, error) {
	// Validate the provided Git configuration.
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid git config: %w", err)
	}

	client := github.NewClient(nil)
	// Valid  Personal Access Token is required if performing actions on a private repository.
	// If no token is provided, the client will be unauthenticated.
	if cfg.PersonalAccessToken != "" {
		client = github.NewClient(nil).WithAuthToken(cfg.PersonalAccessToken)
		logger.Info().Msg("Performing Git operations with Authenticated GitHub Client")
	} else {
		logger.Info().Msg("Performing Git operations with Unauthenticated GitHub Client")
	}

	// Create and return a new Provider with the authenticated GitHub client.
	return &Provider{
		logger: logger,
		Client: client,
	}, nil
}

// PullRequestConfig holds configuration options for creating a pull request.
type PullRequestConfig struct {
	CurrentBranch string // The current branch for the pull request.
	TargetBranch  string // The target branch for the pull request.
	Title         string // The title of the pull request.
	Body          string // The body/description of the pull request.
	Draft         bool   // Indicates whether the PR should be created as a draft.
	Issue         int    // Optional issue number to associate with the PR.
}

// IsValid checks if the pull request configuration is valid.
// It ensures that TargetBranch, Title, and Body are not empty.
func (cfg PullRequestConfig) IsValid() error {
	if cfg.CurrentBranch == "" {
		return fmt.Errorf("pull request config error: current branch is required")
	}
	if cfg.TargetBranch == "" {
		return fmt.Errorf("pull request config error: target branch is required")
	}
	if cfg.Title == "" && cfg.Issue == 0 {
		return fmt.Errorf("pull request config error: title or issue number is required")
	}
	if cfg.Title != "" && cfg.Issue != 0 {
		return fmt.Errorf("pull request config error: title and issue number cannot both be provided")
	}
	if cfg.Body == "" {
		return fmt.Errorf("pull request config error: body is required")
	}
	return nil
}

// CreatePullRequest creates a new pull request on GitHub using the provided configuration.
// It validates the configuration, retrieves repository metadata, pushes the current branch,
// and makes an API call to create the PR.
//
// repoOwner is the GitHub repository owner (organization or username).
// Returns the created pull request on success.
func (p *Provider) CreatePullRequest(ctx context.Context, repoOwner string, cfg PullRequestConfig) (*github.PullRequest, error) {
	// Validate the pull request configuration.
	if err := cfg.IsValid(); err != nil {
		return nil, fmt.Errorf("invalid pull request config: %w", err)
	}

	if repoOwner == "" {
		return nil, fmt.Errorf("repo owner is required")
	}

	// Retrieve the current repository name.
	repoName, err := p.getCurrentRepoName()
	if err != nil {
		return nil, fmt.Errorf("failed to get repo name: %w", err)
	}

	// Ensure the current branch is pushed to the remote repository.
	err = p.PushBranchToRemote(cfg.CurrentBranch)
	if err != nil {
		return nil, fmt.Errorf("failed to push branch to remote: %w", err)
	}

	// Construct the new pull request payload.
	newPR := &github.NewPullRequest{
		Head:  github.Ptr(cfg.CurrentBranch),
		Base:  github.Ptr(cfg.TargetBranch),
		Body:  github.Ptr(cfg.Body),
		Draft: github.Ptr(cfg.Draft),
	}
	// If a title is provided, include it.
	if cfg.Title != "" {
		newPR.Title = github.Ptr(cfg.Title)
	}
	// If an issue number is provided, include it.
	if cfg.Issue != 0 {
		newPR.Issue = github.Ptr(cfg.Issue)
	}

	// Create the pull request via GitHub's API.
	pr, _, err := p.PullRequests.Create(ctx, repoOwner, repoName, newPR)
	if err != nil {
		p.logger.Error().Err(err).Msg("failed to create pull request")
		return nil, fmt.Errorf("%s: %w\n%s", errPullRequestFailed, err, suggestConfiguringPAT)
	}

	// Log the URL of the created pull request.
	p.logger.Info().Str("url", pr.GetHTMLURL()).Msg("created pull request")
	return pr, nil
}

// UpdatePullRequestBody updates the title and body of an existing pull request.
// repoOwner is the GitHub repository owner (organization or username).
func (p *Provider) UpdatePullRequestBody(ctx context.Context, repoOwner string, number int, title, body string) (*github.PullRequest, error) {
	if repoOwner == "" {
		return nil, fmt.Errorf("repo owner is required")
	}
	if number == 0 {
		return nil, fmt.Errorf("pull request number is required")
	}
	if title == "" {
		return nil, fmt.Errorf("pull request title is required")
	}
	if body == "" {
		return nil, fmt.Errorf("pull request body is required")
	}

	// Retrieve the current repository name.
	repoName, err := p.getCurrentRepoName()
	if err != nil {
		return nil, fmt.Errorf("failed to get repo name: %w", err)
	}

	// Create a pull request object with the new title and body.
	pull := &github.PullRequest{
		Title: github.Ptr(title),
		Body:  github.Ptr(body),
	}

	// Call the Edit method to update the pull request.
	pr, _, err := p.PullRequests.Edit(ctx, repoOwner, repoName, number, pull)
	if err != nil {
		return nil, fmt.Errorf("failed to update pull request: %w", err)
	}

	// Log the URL of the updated pull request.
	p.logger.Info().Str("url", pr.GetHTMLURL()).Msg("updated pull request")
	return pr, nil
}

// GetPRTargetBranch retrieves the target branch of a pull request.
// repoOwner is the GitHub repository owner (organization or username).
func (p *Provider) GetPRTargetBranch(ctx context.Context, repoOwner string, number int) (string, error) {
	if repoOwner == "" {
		return "", fmt.Errorf("repo owner is required")
	}
	repoName, err := p.getCurrentRepoName()
	if err != nil {
		return "", fmt.Errorf("failed to get repo name: %w", err)
	}

	pr, _, err := p.PullRequests.Get(ctx, repoOwner, repoName, number)
	if err != nil {
		return "", fmt.Errorf("failed to get pull request: %w", err)
	}

	ref := pr.GetBase().GetRef()
	if ref == "" {
		return "", fmt.Errorf("pull request target branch is empty")
	}

	return ref, nil
}

// PushBranchToRemote pushes the specified branch to the remote repository using the stored personal access token.
// It constructs and executes the "git push" command.
func (p *Provider) PushBranchToRemote(branchName string) error {
	// Construct the git push command.
	cmd := exec.Command("git", "push", "origin", branchName)

	// Execute the command and capture output.
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to push branch to remote: %w\nOutput: %s", err, string(output))
	}

	// Log success.
	p.logger.Info().Msgf("Branch %s pushed to remote successfully", branchName)
	return nil
}

// Below are command templates and constants used to generate a unified diff.
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
	p.logger.Info().Msgf("Executing git diff command ...")
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

	// Wrap the final output in a markdown diff code block.
	wrappedOutput := fmt.Sprintf("```diff\n%s\n```", string(finalOutput))
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

// getCurrentRepoName returns the name of the current repository.
// It extracts the repository name from the base directory of the worktree.
func (p *Provider) getCurrentRepoName() (string, error) {
	// Get the current working directory.
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Open the Git repository.
	repo, err := git.PlainOpen(dir)
	if err != nil {
		return "", err
	}

	// Access the worktree.
	worktree, err := repo.Worktree()
	if err != nil {
		return "", err
	}

	// Derive and return the repository name from the worktree's root.
	repoName := filepath.Base(worktree.Filesystem.Root())
	return repoName, nil
}

// GetCurrentBranchName returns the name of the current branch in the repository.
// It retrieves the HEAD reference and extracts the branch's short name.
func (p *Provider) GetCurrentBranchName() (string, error) {
	// Get the current directory.
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Open the Git repository.
	repo, err := git.PlainOpen(dir)
	if err != nil {
		return "", err
	}

	// Retrieve the repository's HEAD reference.
	ref, err := repo.Head()
	if err != nil {
		return "", err
	}

	// Extract and return the branch name.
	branchName := ref.Name().Short()
	return branchName, nil
}

// GetCommitsSinceCurrentBranchCreation returns the commit messages since the current branch was created.
// This is a convenience wrapper around GetCommitsSinceBranchCreation that attempts to automatically
// detect the base branch.
func (p *Provider) GetCommitsSinceCurrentBranchCreation() ([]string, error) {
	return p.GetCommitsSinceBranchCreation("")
}

// GetCommitsSinceBranchCreation returns the commit messages since the current branch was created.
// It executes git commands to find the commit where the branch diverged from the specified base branch
// and returns all commit messages from that point as a slice of strings.
// If baseBranch is empty, it will attempt to use main, then master, and finally find the fork point.
func (p *Provider) GetCommitsSinceBranchCreation(baseBranch string) ([]string, error) {
	// Get the current branch
	currentBranch, err := p.GetCurrentBranchName()
	if err != nil {
		return nil, fmt.Errorf("failed to get current branch name: %w", err)
	}

	var baseCommit []byte
	var cmd *exec.Cmd

	if baseBranch == "" {
		// Try to detect the base branch automatically
		baseBranches := []string{"main", "master", "develop"}
		for _, branch := range baseBranches {
			cmd = exec.Command("git", "merge-base", branch, currentBranch)
			baseCommit, err = cmd.Output()
			if err == nil {
				p.logger.Info().Msgf("Detected %s as the base branch", branch)
				break
			}
		}

		// If we couldn't find a common ancestor with standard branches, try to find the fork point
		if err != nil {
			p.logger.Info().Msg("Standard base branches not found, finding fork point...")
			cmd = exec.Command("git", "rev-parse", currentBranch)
			currentCommit, err := cmd.Output()
			if err != nil {
				return nil, fmt.Errorf("failed to get current branch commit: %w", err)
			}

			// Use git rev-list to find the fork point
			cmd = exec.Command("bash", "-c", fmt.Sprintf("git rev-list --boundary %s..HEAD | grep ^- | cut -c2-", string(bytes.TrimSpace(currentCommit))))
			baseCommit, err = cmd.Output()
			if err != nil {
				return nil, fmt.Errorf("failed to find fork point: %w", err)
			}
		}
	} else {
		// Use the specified base branch
		cmd = exec.Command("git", "merge-base", baseBranch, currentBranch)
		baseCommit, err = cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("failed to find merge base with %s: %w", baseBranch, err)
		}
	}

	// Trim any whitespace/newlines from the commit hash
	baseCommitHash := string(bytes.TrimSpace(baseCommit))

	// Get all commits between the base commit and HEAD
	cmd = exec.Command("git", "log", "--pretty=format:%s", baseCommitHash+"..HEAD")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get commit messages: %w", err)
	}

	// If there are no commits, return an empty slice
	if len(output) == 0 {
		p.logger.Info().Msg("No commits found since branch creation")
		return []string{}, nil
	}

	// Split the output by newlines to get individual commit messages
	commitMessages := bytes.Split(output, []byte("\n"))

	// Convert byte slices to strings
	result := make([]string, len(commitMessages))
	for i, msg := range commitMessages {
		result[i] = string(msg)
	}

	p.logger.Info().Msgf("Found %d commits since branch creation", len(result))
	return result, nil
}
