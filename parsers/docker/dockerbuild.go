package docker

// The Dockerfile parser is optimized to find and update FROM statements

import (
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/vsoch/uptodate/config"
	"github.com/vsoch/uptodate/parsers"
	"github.com/vsoch/uptodate/parsers/spack"
	"github.com/vsoch/uptodate/utils"
)

// DockerBuildParser holds one or more Docker Builds
type DockerBuildParser struct{}

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

// parseContainerBuildArg parses a spack build arg
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
		versions := GetVersions(buildarg.Name, buildarg.Filter, buildarg.StartAt, buildarg.Skips, buildarg.Includes)
		newVar := parsers.BuildVariable{Name: key, Values: versions}
		vars = append(vars, newVar)

	}
	return vars

}

// parseSpackBuildArg parses a spack build arg
func parseSpackBuildArg(key string, buildarg config.BuildArg) []parsers.BuildVariable {

	// Get versions for current spack package
	packageUrl := "https://spack.github.io/packages/data/packages/" + buildarg.Name + ".json"
	response := utils.GetRequest(packageUrl)

	// The response gets parsed into a spack package
	pkg := spack.SpackPackage{}
	err := json.Unmarshal([]byte(response), &pkg)
	if err != nil {
		log.Fatalf("Issue unmarshalling %s\n", packageUrl)
	}

	// Get versions based on user preferences
	versions := pkg.GetVersions(buildarg.Filter, buildarg.StartAt, buildarg.Skips, buildarg.Includes)
	newVar := parsers.BuildVariable{Name: key, Values: versions}
	vars := []parsers.BuildVariable{newVar}
	return vars
}

// Entrypoint to parse one or more Docker build matrices
func (s *DockerBuildParser) Parse(path string) error {

	// Find config files in path and don't allow prefixes
	paths, _ := utils.RecursiveFind(path, "uptodate.yaml", false)

	// Look at each found path, parse into build matrix
	for _, subpath := range paths {
		conf := config.Load(subpath)

		// We must have a DockerBuild to continue! This checks against an empty one
		if reflect.DeepEqual(conf.DockerBuild, config.DockerBuild{}) {
			log.Printf("dockerbuild section not detected in config, skipping %s\n", subpath)
			continue
		}

		// Prepare lists of values to create a matrix over
		vars := []parsers.BuildVariable{}

		for key, buildarg := range conf.DockerBuild.BuildArgs {

			// If it has a type, it either is that type, or we map to another type
			if buildarg.Type == "container" {
				result := parseContainerBuildArg(key, buildarg)
				vars = append(vars, result...)
			} else if buildarg.Type == "spack" {
				result := parseSpackBuildArg(key, buildarg)
				vars = append(vars, result...)
			} else {
				result := parseBuildArg(key, buildarg)
				vars = append(vars, result...)
			}
		}

		// Create build matrix and format into build results
		matrix := GetBuildMatrix(vars)
		results := []parsers.BuildResult{}

		// Find Dockerfile in subpath
		dirname := filepath.Dir(subpath)
		dockerfiles, _ := utils.RecursiveFind(dirname, "Dockerfile", true)

		// We need a new build for each Dockerfile found (hopefully not many)
		for _, dockerfile := range dockerfiles {
			for _, entry := range matrix {

				// Generate a suggested command, assuming using the dockerfile in its directory
				command := generateBuildCommand(entry, dockerfile)
				description := generateBuildDescription(entry, dockerfile)
				fmt.Println(command)
				newResult := parsers.BuildResult{BuildArgs: entry, CommandPrefix: command, Description: description, Filename: dockerfile, Parser: "dockerbuild", Name: subpath}
				results = append(results, newResult)
			}
		}

		// Parse into json
		outJson, _ := json.Marshal(results)

		// If we are running in a GitHub Action, set the outputs

		if utils.IsGitHubAction() {
			fmt.Printf("::set-output name=dockerbuild_matrix::%s\n", string(outJson))
		} else {
			fmt.Printf("%s\n", string(outJson))
		}
	}
	return nil
}

// GetBuildMatrix generates a build matrix, across all variable options
func GetBuildMatrix(vars []parsers.BuildVariable) []map[string]string {

	// The final result is a list of key value pairs
	results := []map[string]string{}

	for _, buildvar := range vars {
		newResults := getBuildMatrix(buildvar.Name, buildvar.Values, results)
		results = append(results, newResults...)
	}
	return results
}

// generateBuildCommand will generate a build command for a given Dockerfile and buildards
func generateBuildCommand(buildargs map[string]string, dockerfile string) string {

	// The build should be relative to where the Dockerfile is
	filename := filepath.Base(dockerfile)

	// Start the command (use environment variable for name)
	command := "docker build -f " + filename

	// Add each buildarg
	for key, value := range buildargs {
		command += " --build-arg " + key + "=" + value
	}
	return command
}

// generateBuildDescription is useful so the build has a human readable string
func generateBuildDescription(buildargs map[string]string, dockerfile string) string {

	// Assume for now the Dockerfile directory is an identifier
	dirname := filepath.Dir(dockerfile)

	// Start the command (use environment variable for name)
	description := dirname

	// Add each buildarg
	for key, value := range buildargs {
		description += " " + key + ":" + value
	}
	return description
}

// getBuildMatrix is a helper function to grow a list of maps with each set of params
func getBuildMatrix(newkey string, values []string, previous []map[string]string) []map[string]string {

	// Special case when no lists yet - we need to return a list of maps with all versions
	if len(previous) == 0 {
		for _, value := range values {
			entry := make(map[string]string)
			entry[newkey] = value
			previous = append(previous, entry)
		}
		return previous
	}

	updated := []map[string]string{}

	// Add each value to each existing
	for _, value := range values {
		for _, entry := range previous {
			entry[newkey] = value
			updated = append(updated, entry)
		}
	}
	return updated
}
