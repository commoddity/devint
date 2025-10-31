package git

import (
	"context"
	"fmt"
	stdlog "log"
	"os"
	"path/filepath"
	"time"

	"github.com/google/go-github/v69/github"
	"github.com/spf13/cobra"

	"github.com/commoddity/devint/config"
	llmCfg "github.com/commoddity/devint/config/llm"
	gitPkg "github.com/commoddity/devint/git"
	"github.com/commoddity/devint/log"
)

// PR number flag
var prNumber int

func init() {
	// Initialize flags.
	summarizeprCmd.Flags().IntVarP(&prNumber, "pr-number", "p", 0, "Pull request number to summarize. [REQUIRED]")
	summarizeprCmd.Flags().StringVarP(&repoOwnerOverride, "repo-owner", "r", "", "GitHub repository owner override. If set, overrides the repo_owner from config. [OPTIONAL]")
	// Initialize LLM-related flags.
	summarizeprCmd.Flags().StringVarP(&llmProviderOverride, "provider-override", "P", "", "LLM provider override. If set the default provider in the config will be overridden. [OPTIONAL]")
	summarizeprCmd.Flags().StringVarP(&llmModelOverride, "model-override", "m", "", "LLM model override. If set the default model in the config will be overridden. [OPTIONAL]")
}

// summarizeprCmd represents the summarizepr command
var summarizeprCmd = &cobra.Command{
	Use:   "summarizepr",
	Short: "Generate an educational summary of a pull request using LLM and save it to a markdown file.",
	Long: `Generate an educational summary of a pull request using LLM and save it to a markdown file.

This command fetches a pull request by number, retrieves its diff, sends it to an LLM
to generate an educational summary explaining the rationale and Django/Python best practices,
and saves a formatted markdown summary to the directory specified by pr_summary_output_dir
in your config file.

The summary includes:
  - PR title and number
  - PR URL and metadata
  - PR description/body
  - LLM-generated educational summary (explaining rationale and best practices)

Flags:
  --pr-number (-p)      : Pull request number to summarize (required).
  --repo-owner (-r)     : GitHub repository owner override (optional).
  --provider-override (-P): LLM provider override (optional).
  --model-override (-m) : LLM model override (optional).`,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize logger.
		logger := log.NewJSONLogger()

		// Validate required flags.
		if prNumber == 0 {
			stdlog.Fatalf("PR number is required. Use --pr-number (-p) flag.")
		}

		// Load configuration from the config YAML file.
		cfg, err := config.LoadConfig()
		if err != nil {
			stdlog.Fatalf("failed to load config: %v", err)
		}

		// Check if pr_summary_output_dir is set.
		if cfg.Git.PRSummaryOutputDir == "" {
			stdlog.Fatalf("pr_summary_output_dir is required in git_config. Set it using 'devint config' or in your config file.")
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
		gitProvider, err := gitPkg.NewGitProvider(logger, *cfg.Git)
		if err != nil {
			stdlog.Fatalf("failed to create git provider: %v", err)
		}

		logger = logger.With(
			"pr_number", prNumber,
			"repo_owner", repoOwner,
		)

		logger.Info("✅ Initialization successful. Fetching PR information...")

		// Get pull request details.
		pr, err := gitProvider.GetPullRequest(cmd.Context(), repoOwner, prNumber)
		if err != nil {
			stdlog.Fatalf("failed to get pull request: %v", err)
		}

		// Get PR diff.
		diff, err := gitProvider.GetPRDiff(cmd.Context(), prNumber)
		if err != nil {
			stdlog.Fatalf("failed to get PR diff: %v", err)
		}

		// Initialize LLM provider for generating summary.
		providerFlags := getProviderFlags()
		llmProvider, err := llmCfg.NewLLMProvider(logger, cfg.LLMs, providerFlags...)
		if err != nil {
			stdlog.Fatalf("failed to get LLM provider: %v", err)
		}

		// Build the prompt with PR information and diff.
		prompt := buildSummaryPrompt(pr, diff)

		// Sanitize the entire prompt after string interpolation to catch any company names
		// or other sensitive information that may have been introduced during formatting.
		sanitizedPrompt := gitProvider.SanitizePrompt(prompt)

		// Get prompt flags based on any model override.
		promptFlags := getPromptFlags()
		// Send the sanitized prompt to the LLM provider to generate summary.
		llmSummary, err := llmProvider.SendPrompt(context.Background(), sanitizedPrompt, promptFlags...)
		if err != nil {
			stdlog.Fatalf("failed to generate summary from LLM: %v", err)
		}

		// Build the markdown summary with LLM-generated explanation.
		summary := buildPRSummary(pr, llmSummary)

		// Ensure output directory exists.
		outputDir := cfg.Git.PRSummaryOutputDir
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			stdlog.Fatalf("failed to create output directory: %v", err)
		}

		// Generate filename based on PR number and date with timestamp.
		filename := fmt.Sprintf("pr-%d-summary-%s.md", prNumber, time.Now().Format("2006-01-02-150405"))
		filepath := filepath.Join(outputDir, filename)

		// Write the summary to file.
		if err := os.WriteFile(filepath, []byte(summary), 0644); err != nil {
			stdlog.Fatalf("failed to write summary file: %v", err)
		}

		logger.Info("✅ PR summary saved successfully", "filepath", filepath)
		fmt.Printf("✅ PR summary saved successfully!\n📄 File: %s\n", filepath)
	},
}

/*--------- Prompt Construction ---------*/

// summaryPromptTemplate provides the instructions for the LLM to generate an educational PR summary.
// It guides the LLM to explain the rationale and Django/Python best practices in a concise format.
const summaryPromptTemplate = `You are helping a senior level engineer learn about a pull request.
The engineer is familiar with Go and Typescript but is relatively new to Python and Django.

Please analyze the following pull request and generate a brief, concise summary (2-4 paragraphs max) that:
1. Explains the rationale behind the changes - why were these modifications made?
2. Highlights key Django and Python best practices demonstrated
3. Points out notable improvements or optimizations.
4. Points out potential pitfalls or gotchas.
5. Looks for any obvious bugs or runtime errors that may be introduced by the changes.

- Keep the summary brief and easily readable.
- When explaining code changes, include the code being explained in a block above the explanation.
- Start the summary with a bullet point list of changed files and models (if any).

PR Information:
- Title: %s
- Description: %s

Diff:

%s`

// buildSummaryPrompt constructs the LLM prompt with PR details and diff.
func buildSummaryPrompt(pr *github.PullRequest, diff string) string {
	title := pr.GetTitle()
	if title == "" {
		title = "Untitled"
	}

	body := pr.GetBody()
	if body == "" {
		body = "*No description provided.*"
	}

	return fmt.Sprintf(summaryPromptTemplate, title, body, diff)
}

// buildPRSummary constructs a markdown summary from PR details and LLM-generated explanation.
func buildPRSummary(pr *github.PullRequest, llmSummary string) string {
	title := pr.GetTitle()
	if title == "" {
		title = "Untitled"
	}

	body := pr.GetBody()
	if body == "" {
		body = "*No description provided.*"
	}

	url := pr.GetHTMLURL()
	number := pr.GetNumber()
	state := pr.GetState()
	author := pr.GetUser().GetLogin()
	createdAt := pr.GetCreatedAt().Format("2006-01-02 15:04:05")

	return fmt.Sprintf(`# Pull Request #%d: %s

## 📋 Metadata

- **Status**: %s
- **Author**: @%s
- **Created**: %s
- **URL**: %s

## 📝 Description

%s

## 💡 Summary & Analysis

%s
`, number, title, state, author, createdAt, url, body, llmSummary)
}
