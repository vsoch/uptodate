package utils

import (
	"os"
)

// Are we running in a GitHub Action?
func IsGitHubAction() bool {
	_, inGitHub := os.LookupEnv("GITHUB_EVENT_NAME")

	return inGitHub
}
