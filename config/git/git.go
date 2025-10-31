package git

import (
	"errors"
	"regexp"
)

var tokenPattern = regexp.MustCompile(`^(ghp_|gho_|ghu_|ghs_|ghr_|github_pat_)[a-zA-Z0-9_]{36,}$`)

var (
	errGitConfigMissing           = errors.New("config error: git config is missing")
	errPersonalAccessTokenMissing = errors.New("config error: personal access token is missing")
	errInvalidPersonalAccessToken = errors.New("config error: personal access token is invalid")
)

type Config struct {
	// Valid Personal Access Token for GitHub. Required if performing actions on a private repository.
	// Should have at least `write` scope for `repo`.
	PersonalAccessToken string `yaml:"personal_access_token"`
	// GitHub repository owner (organization or username) for pull request operations.
	RepoOwner string `yaml:"repo_owner"`
}

func (c *Config) Validate() error {
	if c == nil {
		return errGitConfigMissing
	}
	if c.PersonalAccessToken == "" {
		return errPersonalAccessTokenMissing
	}
	if !tokenPattern.MatchString(c.PersonalAccessToken) {
		return errInvalidPersonalAccessToken
	}
	return nil
}
