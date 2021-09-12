package spack

// The spack Parser can parse Json from the spack packages repository

import (
	"github.com/vsoch/uptodate/parsers"
	"sort"
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
func (s *SpackPackage) GetVersions(filters []string, startAtVersion string, endAtVersion string, skipVersions []string, includeVersions []string) []string {

	// Sort versions from earliest to latest
	contenders := []string{}
	for _, version := range s.Versions {
		contenders = append(contenders, version.Name)
	}

	// Sort from least to greatest
	sort.Sort(sort.Reverse(sort.StringSlice(contenders)))

	return parsers.GetVersions(contenders, filters, startAtVersion, endAtVersion, skipVersions, includeVersions)
}
