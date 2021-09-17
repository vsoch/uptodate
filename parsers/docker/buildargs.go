package docker

// The Dockerfile parser is optimized to find and update FROM statements

import (
	"log"
	"strings"

	"github.com/vsoch/uptodate/config"
	"github.com/vsoch/uptodate/parsers"
	"github.com/vsoch/uptodate/parsers/spack"
)

// parseBuildArg parses a standard build arg
func parseBuildArg(key string, buildarg config.BuildArg) []parsers.BuildVariable {

	// We will return a list of BuildVariable
	vars := []parsers.BuildVariable{}

	// The values can be versions or values (or both I suppose)
	if len(buildarg.Values) > 0 {
		buildvar := parsers.BuildVariable{Name: key, Values: buildarg.Values}
		vars = append(vars, buildvar)
	}

	if len(buildarg.Versions) > 0 {
		buildvar := parsers.BuildVariable{Name: key, Values: buildarg.Versions}
		vars = append(vars, buildvar)
	}

	return vars
}

// parseContainerBuildArg parses a container build arg
func parseContainerBuildArg(key string, buildarg config.BuildArg) []parsers.BuildVariable {

	// We will return a list of BuildVariable
	vars := []parsers.BuildVariable{}

	// The container is required to have the name
	if buildarg.Name == "" {
		log.Fatalf("A container buildarg requires a name: %s\n", buildarg)
	}

	// If the name has a tag, we just update the version. No further parsing
	if strings.Contains(buildarg.Name, ":") {
		fromValue := []string{buildarg.Name}
		update := UpdateFrom(fromValue)
		newVar := parsers.BuildVariable{Name: key, Values: []string{update.Updated}}
		vars = append(vars, newVar)

		// Otherwise we want to be generating a list of tags (versions)
	} else {
		versions := GetVersions(buildarg.Name, buildarg.Filter, buildarg.StartAt, buildarg.EndAt,
			buildarg.Skips, buildarg.Includes)
		newVar := parsers.BuildVariable{Name: key, Values: versions}
		vars = append(vars, newVar)

	}
	return vars

}

// parseSpackBuildArg parses a spack build arg
func parseSpackBuildArg(key string, buildarg config.BuildArg) []parsers.BuildVariable {

	// Get versions for current spack package
	pkg := spack.GetSpackPackage(buildarg.Name)

	// Get versions based on user preferences
	versions := pkg.GetVersions(buildarg.Filter, buildarg.StartAt, buildarg.EndAt, buildarg.Skips, buildarg.Includes)
	newVar := parsers.BuildVariable{Name: key, Values: versions}
	vars := []parsers.BuildVariable{newVar}
	return vars
}
