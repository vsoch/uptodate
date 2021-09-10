package parsers

import (
	"github.com/blang/semver/v4"
	"github.com/vsoch/uptodate/utils"
	"regexp"
	"strings"
)

// A Result object will store a path to some file that was changed, and
// an identifier for the parser, and some identifier for the changed file

type Result struct {
	Name       string `json:"name,omitempty"`
	Filename   string `json:"filename,omitempty"`
	Parser     string `json:"parser,omitempty"`
	Identifier string `json:"id,omitempty"`
}

// A BuildResult needs more information (e.g., versions) to be given to a build matrix
type BuildResult struct {
	Name          string            `json:"name,omitempty"`
	ContainerName string            `json:"container_name,omitempty"`
	Filename      string            `json:"filename,omitempty"`
	Parser        string            `json:"parser,omitempty"`
	BuildArgs     map[string]string `json:"buildargs,omitempty"`
	CommandPrefix string            `json:"command_prefix,omitempty"`
	Description   string            `json:"description,omitempty"`
}

// BuildVariable holds a key (name) and one or more values to parameterize over
type BuildVariable struct {
	Name   string
	Values []string
}

// VersionRegex matches a major and minor, optional third group
var VersionRegex = "[0-9]+[.][0-9]+(?:[.][0-9]+)?"
var SemverRegex = "^[0-9]+[.][0-9]+[.][0-9]$"

func SortVersions(contenders []string) []string {

	// We need to sort the versions
	svers := []semver.Version{}
	for _, version := range contenders {
		v, err := semver.Parse(version)
		if err == nil {
			svers = append(svers, v)
		}
	}
	semver.Sort(svers)

	// Convert to string
	versions := []string{}
	for _, v := range svers {
		version := v.String()
		versions = append(versions, version)
	}
	return versions
}

func HasSemanticVersions(contenders []string) bool {

	// Do we have any semantic versions?
	haveSemver := false
	isSemver, _ := regexp.Compile(SemverRegex)

	// If we do, we can attempt a custom sort
	for _, version := range contenders {
		if isSemver.MatchString(version) {
			haveSemver = true
		}
	}
	return haveSemver
}

func GetVersions(contenders []string, filters []string, startAtVersion string, endAtVersion string, skipVersions []string, includeVersions []string) []string {

	// Final list of versions we will provide
	versions := []string{}

	// Do we have semantic versions? If so, this is the better sort
	if HasSemanticVersions(contenders) {
		contenders = SortVersions(contenders)
	}

	// We look for tags based on filters (this is an OR between them)
	filter := "(" + strings.Join(filters, "|") + ")"
	isVersionRegex, _ := regexp.Compile(filter)

	// Also don't add until we hit the start at version, given defined
	doAdd := true
	if startAtVersion != "" {
		doAdd = false
	}

	// The tags should already be sorted
	for _, version := range contenders {

		// If it's in the list to include, include no matter what
		if utils.IncludesString(version, includeVersions) {
			versions = append(versions, version)
			continue
		}

		// Have we hit the requested start version, and can add now?
		if startAtVersion != "" && startAtVersion == version && !doAdd {
			doAdd = true
		}

		// Is the tag in the list to skip?
		if utils.IncludesString(version, skipVersions) {
			continue
		}

		// If we are adding, great! Add here to our list
		if doAdd && isVersionRegex.MatchString(version) {
			versions = append(versions, version)
		}

		// If we have an endat version and we reached it, break
		if endAtVersion != "" && endAtVersion == version {
			break
		}
	}
	return versions
}
