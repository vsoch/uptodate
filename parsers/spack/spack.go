package spack

// The spack Parser can parse Json from the spack packages repository

import (
	"github.com/vsoch/uptodate/utils"
	"regexp"
	"sort"
	"strings"
)

// A SpackAlias has a name and alias_for
type SpackAlias struct {
	Name     string `json:"name"`
	AliasFor string `json:"alias_for"`
}

type SpackVersion struct {
	Name   string `json:"name"`
	Sha256 string `json:"sha256"`
}

type SpackConflict struct {
	Name        string `json:"name"`
	Spec        string `json:"spec"`
	Description string `json:"description"`
}

// A SpackPackage matches the format of spack.github.io/packages/data/packages/<package>.json
type SpackPackage struct {
	Name         string       `json:"name"`
	Aliases      []SpackAlias `json:"aliases"`
	Versions     []SpackVersion
	BuildSystem  string            `json:"build_system"`
	Conflicts    []SpackConflict   `json:"conflicts"`
	Variants     []SpackVariant    `json:"variants"`
	Homepage     string            `json:"homepage"`
	Patches      []SpackPatch      `json:"patches"`
	Maintainers  []string          `json:"maintainers"`
	Resources    interface{}       `json:"resources"`
	Description  string            `json:"description"`
	Dependencies []SpackDependency `json:"dependencies"`
	DependentTo  []SpackDependency `json:"dependent_to"`
}

type SpackPatch struct {
	Owner        string `json:"owner"`
	Sha256       string `json:"sha256"`
	Level        int    `json:"level"`
	WorkingDir   string `json:"working_dir"`
	RelativePath string `json:"relative_path"`
	Version      string `json:"version"`
}

type SpackVariant struct {
	Name        string      `json:"name"`
	Default     interface{} `json:"default"`
	Description string      `json:"description"`
}

type SpackDependency struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Get Versions of a spack package relevant to a set of user preferences
// TODO this logic is too similar to the docker equivalent of GetVersions - should be one function
func (s *SpackPackage) GetVersions(filters []string, startAtVersion string, skipVersions []string, includeVersions []string) []string {

	// Final list of versions we will provide
	versions := []string{}
	contenders := []string{}

	// We look for tags based on filters (this is an OR between them)
	filter := "(" + strings.Join(filters, "|") + ")"
	isVersionRegex, _ := regexp.Compile(filter)

	// Also don't add until we hit the start at version, given defined
	doAdd := true
	if startAtVersion != "" {
		doAdd = false
	}

	// We will need to sort from earliest to latest
	for _, version := range s.Versions {
		contenders = append(contenders, version.Name)
	}

	// Sort from least to greatest
	sort.Sort(sort.Reverse(sort.StringSlice(contenders)))

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
	}
	return versions
}
