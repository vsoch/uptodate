package docker

// Base functions for docker used by different updaters

import (
	"fmt"
	"sort"
	"strings"

	lookout "github.com/alecbcs/lookout/update"
	"github.com/vsoch/uptodate/parsers"
	"github.com/vsoch/uptodate/utils"
)

// GetVersions of existing container within user preferences
func GetVersions(container string, filters []string, startAtVersion string, endAtVersion string,
	skipVersions []string, includeVersions []string) []string {

	// Get tags for current container image
	tagsUrl := "https://crane.ggcr.dev/ls/" + container
	response := utils.GetRequest(tagsUrl)
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
