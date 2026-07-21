package cmd

import (
	"os"

	"github.com/google/go-github/v59/github"
)

// getGitHubClient initializes and returns a GitHub API client.
// It automatically applies an authentication token if the GITHUB_TOKEN environment
// variable is set, otherwise returning an unauthenticated client subject to stricter
// rate limits.
func getGithubClient() *github.Client {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return github.NewClient(nil)
	}

	return github.NewClient(nil).WithAuthToken(token)
}
