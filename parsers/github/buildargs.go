package github

import (
	"fmt"
	"strings"

	"github.com/vsoch/uptodate/parsers"
)

// UpdateCommitBuildArg will update a github commit build arg
// Should be called from docker.go UpdateArg to ensure intiial checks
// ARG uptodate_github_commit_<org>__<name>__branch=<release-tag>
func UpdateCommitBuildArg(values []string) parsers.Update {

	// This is the full argument with =
	arg := values[0]

	// Keep the original for later comparison
	original := strings.Join(values, " ")

	// We will return an update, empty if none
	update := parsers.Update{}

	// Split into buildarg name and value
	parts := strings.SplitN(arg, "=", 2)
	name := parts[0]

	fmt.Printf("Found github commit build arg prefix %s\n", arg)
	name = strings.Replace(name, "uptodate_github_commit_", "", 1)

	// The repository name must be separated by __
	if strings.Count(name, "__") != 2 {
		fmt.Printf("Cannot find double underscore to separate org from repo name, and then branch: %s", name)
		return update
	}
	orgRepoBranch := strings.SplitN(name, "__", 3)
	org := orgRepoBranch[0]
	repo := orgRepoBranch[1]
	branch := orgRepoBranch[2]

	// Organization __ Repository
	if org == "" || repo == "" || branch == "" {
		fmt.Printf("Org (%s), repository (%s), or branch (%s) is empty, cannot parse.", org, repo, branch)
		return update
	}

	repository := org + "/" + repo
	commits := GetCommits(repository, branch)

	// The first in the list is the newest commit
	if len(commits) == 0 {
		fmt.Printf("%s has no commits, cannot update.", repository)
		return update
	}

	commit := commits[0]
	updated := parts[0] + "=" + commit.SHA

	// Add original content back
	for _, extra := range values[1:] {
		updated += " " + extra
	}

	// If the updated version is different from the original, update
	if updated != original {
		update = parsers.Update{Original: original, Updated: updated}
	} else {
		fmt.Println("No difference between:", updated, original)
	}
	return update

}

// UpdateReleaseBuildArg will update a github release build arg
// Should be called from docker.go UpdateArg to ensure intiial checks
// ARG uptodate_github_release_<org>__<name>=<release-tag>
func UpdateReleaseBuildArg(values []string) parsers.Update {

	// This is the full argument with =
	arg := values[0]

	// Keep the original for later comparison
	original := strings.Join(values, " ")

	// Split into buildarg name and value
	parts := strings.SplitN(arg, "=", 2)
	name := parts[0]

	// We will return an update, empty if none
	update := parsers.Update{}

	fmt.Printf("Found github release build arg prefix %s\n", arg)
	name = strings.Replace(name, "uptodate_github_release_", "", 1)

	// The repository name must be separated by __
	if !strings.Contains(name, "__") {
		fmt.Printf("Cannot find double underscore to separate org from repo name: %s", name)
		return update
	}
	orgRepo := strings.SplitN(name, "__", 2)
	org := orgRepo[0]
	repo := orgRepo[1]

	// Organization __ Repository
	if org == "" || repo == "" {
		fmt.Printf("Org (%s) or repository (%s) is empty, cannot parse.", org, repo)
		return update
	}
	repository := org + "/" + repo
	releases := GetReleases(repository)

	// The first in the list is the newest release
	if len(releases) == 0 {
		fmt.Printf("%s has no releases, cannot update.", repository)
		return update
	}
	release := releases[0]
	updated := parts[0] + "=" + release.Name

	// Add original content back
	for _, extra := range values[1:] {
		updated += " " + extra
	}

	// If the updated version is different from the original, update
	if updated != original {
		update = parsers.Update{Original: original, Updated: updated}
	} else {
		fmt.Println("No difference between:", updated, original)
	}

	return update
}
