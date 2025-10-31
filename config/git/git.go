package git

import (
	"errors"
	"regexp"
)

var tokenPattern = regexp.MustCompile(`^(ghp_|gho_|ghu_|ghs_|ghr_|github_pat_)[a-zA-Z0-9_]{36,}$`)

var (
	errGitConfigMissing           = errors.New("config error: git config is missing")
	errInvalidPersonalAccessToken = errors.New("config error: personal access token is invalid")
)

type Config struct {
	// Valid Personal Access Token for GitHub. Required if performing actions on a private repository.
	// Should have at least `write` scope for `repo`.
	PersonalAccessToken string `yaml:"personal_access_token"`
	// GitHub repository owner (organization or username) for pull request operations.
	RepoOwner string `yaml:"repo_owner"`
	// Optional directory path where PR summaries will be saved. If not set, summaries are not saved to disk.
	PRSummaryOutputDir string `yaml:"pr_summary_output_dir"`
}

func (c *Config) Validate() error {
	if c == nil {
		return errGitConfigMissing
	}
	// PersonalAccessToken is now optional - we can fall back to gh auth token
	if c.PersonalAccessToken != "" && !tokenPattern.MatchString(c.PersonalAccessToken) {
		return errInvalidPersonalAccessToken
	}
	return nil
}
