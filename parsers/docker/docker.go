package docker

// Base functions for docker used by different updaters

import (
	"fmt"
	"sort"
	"strings"

	lookout "github.com/alecbcs/lookout/update"
	"github.com/vsoch/uptodate/parsers"
	"github.com/vsoch/uptodate/parsers/github"
	"github.com/vsoch/uptodate/parsers/spack"
	"github.com/vsoch/uptodate/utils"
)

// GetVersions of existing container within user preferences
func GetVersions(container string, filters []string, startAtVersion string, endAtVersion string,
	skipVersions []string, includeVersions []string) []string {

	// Get tags for current container image
	tagsUrl := "https://crane.ggcr.dev/ls/" + container
	response := utils.GetRequest(tagsUrl, map[string]string{})
	tags := strings.Split(response, "\n")
	sort.Sort(sort.StringSlice(tags))

	return parsers.GetVersions(tags, filters, startAtVersion, endAtVersion, skipVersions, includeVersions)
}

// UpdateFrom updates a single From, and returns an Update
func UpdateFrom(fromValue []string) Update {

	// We will return an update, empty if none
	update := Update{}

	// This is the full container name, e.g., ubuntu:16.04
	container := fromValue[0]

	// Keep the original for later comparison
	original := strings.Join(fromValue, " ")

	// Variable statements we can't reliably update
	isVariable := strings.Contains(container, "$")
	if isVariable {
		return update
	}

	// We want to keep track of having a hash and/or tag
	hasHash := false
	hasTag := false

	// First remove any digest from the container
	if strings.Contains(container, "@") {
		parts := strings.SplitN(container, "@", 2)
		container = parts[0]
		hasHash = true
	}

	// Now extract any tag from the container
	tag := "latest"
	if strings.Contains(container, ":") {
		parts := strings.SplitN(container, ":", 2)
		container = parts[0]
		tag = parts[1]
		hasTag = true
	} else {
		fmt.Printf("No tag specified for %s, will default to latest.\n", container)
	}

	// If it has a hash but no digest, we can't correctly parse
	if hasHash && !hasTag {
		fmt.Printf("Cannot parse %s, has a hash but no tag, cannot be looked up.\n", container)
		return update
	}

	// Get the updated container hash for the tag
	url := container + ":" + tag
	out, found := lookout.CheckUpdate("docker://" + url)

	if found {
		// Prepare the updated string, the result.Name is digest
		result := *out
		updated := url + "@" + result.Name

		// Add original content back
		for _, extra := range fromValue[1:] {
			updated += " " + extra
		}

		// If the updated version is different from the original, update
		if updated != original {

			// TODO I've never seen a multi-line FROM, but this will need
			// adjustment if one exists to replace a range of lines
			update = Update{Original: original, Updated: updated}

		} else {
			fmt.Println("No difference between:", updated, original)
		}
	}
	return update
}

// UpdateArg updates a build arg that is a known pattern
func UpdateArg(values []string) Update {

	// We will return an update, empty if none
	update := Update{}

	// This is the full argument with =
	arg := values[0]

	// If we don't have an = (no default) we cannot update
	if !strings.Contains(arg, "=") {
		fmt.Printf("Cannot update %s, does not have a default\n", arg)
		return update
	}

	// Keep the original for later comparison
	original := strings.Join(values, " ")

	// Split into buildarg name and value
	parts := strings.SplitN(arg, "=", 2)

	name := parts[0]
	value := parts[1]

	// We can't have empty value
	if value == "" {
		fmt.Printf("Cannot update %s, does not have a value\n", arg)
		return update
	}

	// Determine if it matches spack or Github
	if strings.HasPrefix(name, "uptodate_spack") {
		fmt.Printf("Found spack build arg prefix %s\n", arg)

		name = strings.Replace(name, "uptodate_spack_", "", 1)

		// Get versions for current spack package
		pkg := spack.GetSpackPackage(name)

		// Should be sorted with newest first
		if len(pkg.Versions) > 0 {

			updated := parts[0] + "=" + pkg.Versions[0].Name

			// Add any comments back
			for _, extra := range values[1:] {
				updated += " " + extra
			}

			// If the updated version is different from the original, update
			if updated != original {
				update = Update{Original: original, Updated: updated}

			} else {
				fmt.Println("No difference between:", updated, original)
			}
		}

	} else if strings.HasPrefix(name, "uptodate_github_release") {
		fmt.Printf("Found github release build arg prefix %s\n", arg)
		name = strings.Replace(name, "uptodate_github_release_", "", 1)

		// The repository name must be separated by __
		if !strings.Contains(name, "__") {
			fmt.Printf("Cannot find double underscore to separate org from repo name: %s", name)
			return update
		}
		orgRepo := strings.SplitN(name, "__", 2)

		// Organization __ Repository
		if orgRepo[0] == "" || orgRepo[1] == "" {
			fmt.Printf("Org (%s) or repository (%s) is empty, cannot parse.", orgRepo[0], orgRepo[1])
			return update
		}
		repository := orgRepo[0] + "/" + orgRepo[1]
		fmt.Println(repository)
		releases := github.GetReleases(repository)

		// The first in the list is the newest release
		if len(releases) == 0 {
			fmt.Printf("%s has no releases, cannot update.", repository)
		}
		release := releases[0]

		updated := parts[0] + "=" + release.Name

		// Add original content back
		for _, extra := range values[1:] {
			updated += " " + extra
		}

		// If the updated version is different from the original, update
		if updated != original {
			update = Update{Original: original, Updated: updated}
		} else {
			fmt.Println("No difference between:", updated, original)
		}
	} else if strings.HasPrefix(name, "uptodate_github_commit") {
		fmt.Printf("Found github commit build arg prefix %s\n", arg)
		name = strings.Replace(name, "uptodate_github_commit_", "", 1)

		// The repository name must be separated by __
		if !strings.Contains(name, "__") {
			fmt.Printf("Cannot find double underscore to separate org from repo name: %s", name)
			return update
		}
		orgRepoBranch := strings.SplitN(name, "__", 3)

		// Organization __ Repository
		if orgRepoBranch[0] == "" || orgRepoBranch[1] == "" || orgRepoBranch[2] == "" {
			fmt.Printf("Org (%s), repository (%s), or branch (%s) is empty, cannot parse.", orgRepoBranch[0], orgRepoBranch[1], orgRepoBranch[2])
			return update
		}
		repository := orgRepoBranch[0] + "/" + orgRepoBranch[1]
		branch := orgRepoBranch[2]
		fmt.Println(repository)
		commits := github.GetCommits(repository, branch)

		// The first in the list is the newest commit
		if len(commits) == 0 {
			fmt.Printf("%s has no commits, cannot update.", repository)
		}
		commit := commits[0]

		updated := parts[0] + "=" + commit.SHA

		// Add original content back
		for _, extra := range values[1:] {
			updated += " " + extra
		}

		// If the updated version is different from the original, update
		if updated != original {
			update = Update{Original: original, Updated: updated}
		} else {
			fmt.Println("No difference between:", updated, original)
		}
	}
	return update
}
