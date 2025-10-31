// ---------------------------------------------------------------------------
// File: createpr.go
// Package: git
//
// Purpose:
//   This command automatically generates a GitHub Pull Request (PR) description
//   using a language model (LLM) and then creates a PR on GitHub. It does so by
//   computing a diff from the local git repository, generating a prompt for the LLM,
//   and incorporating the LLM's response into a PR description template. The command
//   also supports a dummy mode where the PR is not created but instead its description
//   is printed and copied to the clipboard.
//
// Features:
//   - Uses LLM to generate a PR description from a git diff.
//   - Supports overriding the default LLM provider and model using flags.
//   - Validates required flags (e.g. a non-empty PR title).
//   - If in dummy mode, does not create the PR and copies the PR description to
//     the clipboard instead.
//   - Logs relevant information with structured JSON logging using slog.
//   - Supports linking an issue number to the PR, and creating the PR as a draft if
//     the title contains "DRAFT".
// ---------------------------------------------------------------------------

package git

import (
	"context"
	"fmt"
	stdlog "log"
	"strings"

	"github.com/spf13/cobra"
	"golang.design/x/clipboard"

	"github.com/commoddity/devint/config"
	llmCfg "github.com/commoddity/devint/config/llm"
	"github.com/commoddity/devint/git"
	"github.com/commoddity/devint/llm"
	"github.com/commoddity/devint/log"
)

// Git config flags
var prTitle string
var issue int
var targetBranch string
var dummy bool
var updatePRNumber int
var repoOwnerOverride string

// LLM config flags
var llmProviderOverride string
var llmModelOverride string

func init() {
	// Initialize Git-related flags.
	createprCmd.Flags().StringVarP(&prTitle, "pr-title", "t", "", "PR title to open the PR with. Should be descriptive and contain relevant tags. Will open a draft PR if the string contains [DRAFT] or [WIP]. [ONE OF THE FOLLOWING FLAGS MUST BE PROVIDED: -t or -i]")
	createprCmd.Flags().IntVarP(&issue, "issue", "i", 0, "Issue number to assign to the PR. [ONE OF THE FOLLOWING FLAGS MUST BE PROVIDED: -t or -i]")
	createprCmd.Flags().StringVarP(&targetBranch, "target-branch", "b", "main", "Target branch to open the PR on. [OPTIONAL, defaults to 'main']")
	createprCmd.Flags().BoolVarP(&dummy, "dummy", "d", false, "Dummy mode. If true, the command will not create a PR but will only print the PR description and copy to clipboard [OPTIONAL, defaults to 'false']")
	createprCmd.Flags().IntVarP(&updatePRNumber, "update", "u", 0, "Update an existing PR with the given pull request number. [OPTIONAL]")
	createprCmd.Flags().StringVarP(&repoOwnerOverride, "repo-owner", "r", "", "GitHub repository owner override. If set, overrides the repo_owner from config. [OPTIONAL]")
	// Initialize LLM-related flags.
	createprCmd.Flags().StringVarP(&llmProviderOverride, "provider-override", "p", "", "LLM provider override. If set the default provider in the config will be overridden. [OPTIONAL]")
	createprCmd.Flags().StringVarP(&llmModelOverride, "model-override", "m", "", "LLM model override. If set the default model in the config will be overridden. [OPTIONAL]")
}

// createprCmd represents the createpr command
var createprCmd = &cobra.Command{
	Use:   "createpr",
	Short: "Automatically generate a PR description and open it on GitHub.",
	Long: `Automatically generate a PR description and open it on GitHub.

This command automatically generates a PR description by using an LLM to summarize a git diff.
It first computes a diff against a target branch (default "main"), builds a prompt incorporating
the diff and the PR title, and sends that prompt to the LLM. The LLM's response is then embedded
into a PR description template that includes a sanity checklist. Finally, the command creates a 
PR on GitHub using the generated description. 

Flags:
  --pr-title (-t)   : PR title. (Exactly one of the following flags must be provided: -t or -i)
  --issue (-i)      : Issue number to assign to the PR. (Exactly one of the following flags must be provided: -t or -i)
  --target-branch (-b): Target branch (default "main").
  --dummy (-d)      : If set, the PR is not created; instead, the description is printed and copied.
  --update (-u)     : Update an existing PR with the given pull request number. Cannot be used with --issue.
  --repo-owner (-r) : GitHub repository owner override. If set, overrides the repo_owner from config.
  --provider-override (-p): Optional LLM provider override.
  --model-override (-m)   : Optional LLM model override.`,
	Run: func(cmd *cobra.Command, args []string) {

		// Initialize logger.
		logger := log.NewJSONLogger()

		// Validate required flags.
		if prTitle == "" && issue == 0 {
			stdlog.Fatalf("PR title or issue number is required")
		}
		if prTitle != "" && issue != 0 {
			stdlog.Fatalf("PR title and issue number cannot both be provided")
		}
		if updatePRNumber != 0 && issue != 0 {
			stdlog.Fatalf("Update PR number and issue number cannot both be provided")
		}

		// Load configuration from the config YAML file.
		cfg, err := config.LoadConfig()
		if err != nil {
			stdlog.Fatalf("failed to load config: %v", err)
		}

		// Determine repo owner: flag override > config > error if neither
		repoOwner := repoOwnerOverride
		if repoOwner == "" {
			repoOwner = cfg.Git.RepoOwner
		}
		if repoOwner == "" {
			stdlog.Fatalf("repo owner is required. Either set repo_owner in git_config (using 'devint config') or provide --repo-owner flag")
		}

		// Initialize the Git provider using loaded Git config.
		gitProvider, err := git.NewGitProvider(logger, *cfg.Git)
		if err != nil {
			stdlog.Fatalf("failed to create git provider: %v", err)
		}

		// Get additional provider flags based on any overrides set.
		providerFlags := getProviderFlags()
		// Initialize the LLM provider with potential provider overrides.
		llmProvider, err := llmCfg.NewLLMProvider(logger, cfg.LLMs, providerFlags...)
		if err != nil {
			stdlog.Fatalf("failed to get LLM provider: %v", err)
		}

		// Retrieve the current branch name.
		currentBranch, err := gitProvider.GetCurrentBranchName()
		if err != nil {
			stdlog.Fatalf("failed to get current branch name: %v", err)
		}

		logger = logger.With(
			"dummy", dummy,
			"current_branch", currentBranch,
			"target_branch", targetBranch,
			"type", "create",
		)

		// If updating an existing PR, get the target branch.
		if updatePRNumber != 0 {
			targetBranch, err = gitProvider.GetPRTargetBranch(context.Background(), repoOwner, updatePRNumber)
			if err != nil {
				stdlog.Fatalf("failed to get PR target branch: %v", err)
			}
			logger = logger.With(
				"target_branch", targetBranch,
				"type", "update",
			)
		}

		logger.Info("✅ Initialization successful. Running command...")

		// Generate a diff from Git against the target branch.
		diff, err := gitProvider.GenerateDiff(context.Background(), targetBranch)
		if err != nil {
			stdlog.Fatalf("failed to generate diff: %v", err)
		}

		// Get the commits on the current branch.
		branchCommits, err := gitProvider.GetCommitsSinceBranchCreation(targetBranch)
		if err != nil {
			stdlog.Fatalf("failed to get branch commits: %v", err)
		}
		branchCommitsStr := strings.Join(branchCommits, "\n")

		// Build the prompt by merging PR title and generated diff.
		prompt := buildPrompt(prTitle, diff, branchCommitsStr)

		// Get prompt flags based on any model override.
		promptFlags := getPromptFlags()
		// Send the prompt to the LLM provider.
		prDescription, err := llmProvider.SendPrompt(context.Background(), prompt, promptFlags...)
		if err != nil {
			stdlog.Fatalf("failed to send prompt: %v", err)
		}

		// --- NEW CODE: Check for TODOs in the diff and append them to the PR description ---
		todoInThisPR, otherTODOs, err := checkForTODOs(diff)
		if err != nil {
			stdlog.Printf("failed to check for TODOs: %v", err)
		}
		if len(todoInThisPR) > 0 {
			todoSection := fmt.Sprintf(todoInThisPRTemplate, strings.Join(todoInThisPR, "\n"))
			prDescription = todoSection + "\n" + prDescription
		}
		if len(otherTODOs) > 0 {
			newTodoSection := fmt.Sprintf(newTODOsTemplate, strings.Join(otherTODOs, "\n"))
			if idx := strings.Index(prDescription, "## 🛠️ Type of change"); idx != -1 {
				prDescription = prDescription[:idx] + newTodoSection + "\n" + prDescription[idx:]
			} else {
				prDescription = newTodoSection + "\n" + prDescription
			}
		}
		// --- END NEW CODE ---

		// In dummy mode, print the PR description and copy it to the clipboard.
		if dummy {
			err := clipboard.Init()
			if err != nil {
				stdlog.Fatalf("failed to initialize clipboard: %v", err)
			}
			fmt.Printf("PR Description:\n\n%s\n", prDescription)
			clipboard.Write(clipboard.FmtText, []byte(prDescription))
			logger.Info("📋 PR description copied to clipboard. PR not created.")
			return
		}

		// If an update PR number is provided, update the pull request.
		if updatePRNumber != 0 {
			pullRequest, err := gitProvider.UpdatePullRequestBody(context.Background(), repoOwner, updatePRNumber, prTitle, prDescription)
			if err != nil {
				stdlog.Fatalf("failed to update pull request: %v", err)
			}
			fmt.Printf("✅ Pull request # %d updated Successfully!\n🌿 Pull Request URL: %s\n", *pullRequest.Number, *pullRequest.HTMLURL)
			return
		}

		// Construct the pull request configuration.
		pullRequestConfig := git.PullRequestConfig{
			CurrentBranch: currentBranch,
			TargetBranch:  targetBranch,
			Title:         prTitle,
			Body:          prDescription,
			Draft:         isDraft(prTitle),
		}
		// If an issue number is provided, add it to the configuration.
		if issue != 0 {
			pullRequestConfig.Issue = issue
		}
		// Create the pull request using the Git provider.
		pullRequest, err := gitProvider.CreatePullRequest(context.Background(), repoOwner, pullRequestConfig)
		if err != nil {
			stdlog.Fatalf("failed to create pull request: %v", err)
		}

		fmt.Printf("✅ Pull request # %d created Successfully!\n🌿 Pull Request URL: %s\n", *pullRequest.Number, *pullRequest.HTMLURL)
	},
}

/*--------- LLM Provider and Prompt Flags ---------*/

// getProviderFlags returns the LLM provider flags, modifying the LLM configuration.
func getProviderFlags() []llmCfg.ProviderFlag {
	var flags []llmCfg.ProviderFlag

	// Append provider override flag if specified.
	if llmProviderOverride != "" {
		flags = append(flags, llmCfg.WithProviderOverride(llmCfg.ProviderType(llmProviderOverride)))
	}

	return flags
}

// getPromptFlags returns the LLM prompt flags, modifying the LLM prompt configuration.
func getPromptFlags() []llm.PromptFlag {
	var flags []llm.PromptFlag

	// Append model override flag if provided.
	if llmModelOverride != "" {
		flags = append(flags, llm.WithLLMModelOverride(llmModelOverride))
	}

	return flags
}

/*--------- Prompt Construction ---------*/

// buildPrompt constructs the LLM prompt by merging the PR title and git diff.
// The prompt follows a specific template to guide the LLM in generating a PR description.
func buildPrompt(prTitle, diff, branchCommits string) string {
	return fmt.Sprintf(promptIntro, git.CombinedDiffCmd, prTitle, branchCommits, diff)
}

// isDraft checks if the PR title contains "DRAFT" (case-insensitive).
// If it does, the PR will be created as a draft.
func isDraft(prTitle string) bool {
	return strings.Contains(strings.ToUpper(prTitle), "DRAFT") ||
		strings.Contains(strings.ToUpper(prTitle), "WIP")
}

/*--------- TODOs ---------*/

// --- NEW FUNCTION: checkForTODOs ---
// checkForTODOs scans the provided diff for added lines containing "TODO_" and,
// if found, captures the complete comment (including any continuations) and returns
// two slices of bullet strings with the full comment and file name.
func checkForTODOs(diff string) (todoInThisPR []string, otherTODOs []string, err error) {
	lines := strings.Split(diff, "\n")
	var currentFile string
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		// Skip markdown code block markers
		if strings.HasPrefix(line, "```") {
			continue
		}
		// Update current file if indicated in diff headers
		if strings.HasPrefix(line, "New File:") {
			currentFile = strings.TrimSpace(strings.TrimPrefix(line, "New File:"))
			continue
		}
		if strings.HasPrefix(line, "+++ b/") {
			currentFile = strings.TrimSpace(strings.TrimPrefix(line, "+++ b/"))
			continue
		}
		// Process added lines (ignore diff metadata)
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			trimmed := strings.TrimSpace(strings.TrimPrefix(line, "+"))
			// Remove comment markers from the initial line
			if strings.HasPrefix(trimmed, "#") {
				trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, "#"))
			} else if strings.HasPrefix(trimmed, "//") {
				trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, "//"))
			}
			if strings.Contains(trimmed, "TODO_") {
				fullComment := trimmed
				// Check subsequent lines for continuation if they are additions with comment markers
				for j := i + 1; j < len(lines); j++ {
					nextLine := lines[j]
					if !strings.HasPrefix(nextLine, "+") || strings.HasPrefix(nextLine, "+++") {
						break
					}
					nextTrimmed := strings.TrimSpace(strings.TrimPrefix(nextLine, "+"))
					// Continue if the line starts with a comment marker
					if strings.HasPrefix(nextTrimmed, "#") || strings.HasPrefix(nextTrimmed, "//") {
						// Remove the comment marker and append
						if strings.HasPrefix(nextTrimmed, "#") {
							nextTrimmed = strings.TrimSpace(strings.TrimPrefix(nextTrimmed, "#"))
						} else if strings.HasPrefix(nextTrimmed, "//") {
							nextTrimmed = strings.TrimSpace(strings.TrimPrefix(nextTrimmed, "//"))
						}
						fullComment += " " + nextTrimmed
						i = j
					} else {
						break
					}
				}
				// Build the bullet without a line number
				bullet := fmt.Sprintf("- %s - %s", fullComment, currentFile)
				if strings.Contains(fullComment, "TODO_IN_THIS_PR") {
					todoInThisPR = append(todoInThisPR, bullet)
				} else {
					otherTODOs = append(otherTODOs, bullet)
				}
			}
		}
	}
	return todoInThisPR, otherTODOs, nil
}

/*--------- LLM Prompt Templates ---------*/

// promptIntro provides the instructions and template structure for the LLM.
// It guides the LLM on how to generate the PR description.
const promptIntro = `Please generate a GitHub PR description from the following diff output. The diff was generated using this command:

		%s

		Where the first 's' is the repo root and the second 's' is the target branch.

		Format:
		- The generated description should be in the following template format.
		- It should output ONLY the template inside the --- BEGIN PR TEMPLATE --- and --- END PR TEMPLATE --- tags.
		- Do NOT include the --- BEGIN PR TEMPLATE --- and --- END PR TEMPLATE --- tags in the output.
		- Use as many Primary and Secondary changes as you feel are valid to capture the changes in the diff.

		--- BEGIN PR TEMPLATE (do not include this line in the output) ---

		## 💡 Summary

		< One line summary >

		### 🛠️ Primary Changes:
		- < core changes # 1 >
		- < core changes # 2 >
		- ...

		### 🧹 Secondary changes:
		- < secondary changes # 1 >
		- < secondary changes # 2 >
		- ...

		--- END PR TEMPLATE (do not include this line in the output) ---

		Focus:
		The PR title should be used to focus the summary and primary/secondary changes as it describes the "why" of the PR.
		Do not simply use it as the Summary but use it as a simple guideline for how to structure the the template.

		PR Title: %s

		The commits on the current branch should also be used to provide context for the PR description and guide the outputted PR description.

		Branch Commits:
		%s

		Considerations:
		- Keep the bullet points concise
		- Escape key terms with backticks
		- Primary changes are the functional changes that are the main focus of the PR
		- Secondary changes include meta or miscellaneous changes (e.g. documentation updates, code cleanup, etc)
		- Limit the number of bullets to 3-5
		- Do not include backticks or markdown formatting in the output

		Diff:

		%s`

const todoInThisPRTemplate = `
## 🚨 TODO_IN_THIS_PR 

Do not merge until these TODOs are resolved:
%s
`

const newTODOsTemplate = `
## 💡 New TODOs 

New TODOs introduced in this PR:
%s
`
