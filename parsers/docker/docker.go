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
func UpdateFrom(fromValue []string) parsers.Update {

	// We will return an update, empty if none
	update := parsers.Update{}

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
			update = parsers.Update{Original: original, Updated: updated}

		} else {
			fmt.Println("No difference between:", updated, original)
		}
	}
	return update
}

// UpdateArg updates a build arg that is a known pattern
func UpdateArg(values []string) parsers.Update {

	// We will return an update, empty if none
	update := parsers.Update{}

	// This is the full argument with =
	arg := values[0]

	// If we don't have an = (no default) we cannot update
	if !strings.Contains(arg, "=") {
		fmt.Printf("Cannot update %s, does not have a default\n", arg)
		return update
	}

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
		return spack.UpdateBuildArg(values)
	} else if strings.HasPrefix(name, "uptodate_github_release") {
		return github.UpdateReleaseBuildArg(values)
	} else if strings.HasPrefix(name, "uptodate_github_commit") {
		return github.UpdateCommitBuildArg(values)
	}
	return update
}
