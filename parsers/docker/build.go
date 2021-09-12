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
	"github.com/vsoch/uptodate/parsers/git"
	"github.com/vsoch/uptodate/parsers/spack"
	"github.com/vsoch/uptodate/utils"
)

// DockerBuildParser holds one or more Docker Builds
type DockerBuildParser struct{}

// Keep track of an original lookup key and the slug for the buildarg
type ContainerNamer struct {
	Slug string
	Key  string
}

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
	packageUrl := "https://spack.github.io/packages/data/packages/" + buildarg.Name + ".json"
	response := utils.GetRequest(packageUrl)

	// The response gets parsed into a spack package
	pkg := spack.SpackPackage{}
	err := json.Unmarshal([]byte(response), &pkg)
	if err != nil {
		log.Fatalf("Issue unmarshalling %s\n", packageUrl)
	}

	// Get versions based on user preferences
	versions := pkg.GetVersions(buildarg.Filter, buildarg.StartAt, buildarg.EndAt, buildarg.Skips, buildarg.Includes)
	newVar := parsers.BuildVariable{Name: key, Values: versions}
	vars := []parsers.BuildVariable{newVar}
	return vars
}

// Entrypoint to parse one or more Docker build matrices
func (s *DockerBuildParser) Parse(path string, changesOnly bool, branch string) error {

	// Find config files in path and don't allow prefixes
	paths, _ := utils.RecursiveFind(path, "uptodate.yaml", false)

	// If we want changed only, honor that
	if changesOnly {

		// Create list of changes (Modify or Add)
		changed := git.GetChangedFilesStrings(path, branch)
		paths = utils.FindOverlap(paths, changed)
	}

	// No updated?
	if len(paths) == 0 {
		fmt.Println("No changes to parse.")
	}

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

		// Keep a record of which variables to use for naming (container vs tag)
		namingLookup := make(map[string][]ContainerNamer)
		namingLookup["container"] = []ContainerNamer{}
		namingLookup["tag"] = []ContainerNamer{}

		// Keys we will skip if not included in matrix
		allowKeys := []string{}

		// If the config already has a matrix, honor it
		var matrix []map[string]string
		if len(conf.DockerBuild.Matrix) > 0 {
			matrix = NewBuildMatrix(conf.DockerBuild.Matrix)

			// We won't build variables not in matrix
			if len(matrix) > 0 {
				firstEntry := matrix[0]
				for key := range firstEntry {
					allowKeys = append(allowKeys, key)
				}
			}

		}

		for key, buildarg := range conf.DockerBuild.BuildArgs {

			// Skip those that aren't in matrix, if matrix predefined
			if len(allowKeys) > 0 && !utils.IncludesString(key, allowKeys) {
				continue
			}

			// identifier is the key and fallback to the name
			namer := ContainerNamer{Key: key, Slug: buildarg.GetKey()}

			// If it has a type, it either is that type, or we map to another type
			if buildarg.Type == "container" {
				result := parseContainerBuildArg(key, buildarg)
				vars = append(vars, result...)
				namingLookup["container"] = append(namingLookup["container"], namer)
			} else if buildarg.Type == "spack" {
				result := parseSpackBuildArg(key, buildarg)
				vars = append(vars, result...)
				namingLookup["tag"] = append(namingLookup["tag"], namer)
			} else {
				result := parseBuildArg(key, buildarg)
				vars = append(vars, result...)
				namingLookup["tag"] = append(namingLookup["tag"], namer)
			}
		}

		// If we don't have the matrix yet, create all possible combinations
		if len(conf.DockerBuild.Matrix) == 0 {
			matrix = GetBuildMatrix(vars)
		}

		// Prepare a list of build results
		results := []parsers.BuildResult{}

		// Find Dockerfile in subpath
		dirnamePath := filepath.Dir(subpath)
		dockerfiles, _ := utils.RecursiveFind(dirnamePath, "Dockerfile", true)
		dirname := filepath.Base(dirnamePath)

		// We need a new build for each Dockerfile found (hopefully not many)
		for _, dockerfile := range dockerfiles {
			for _, entry := range matrix {

				// Generate a suggested command, assuming using the dockerfile in its directory
				command := generateBuildCommand(entry, dockerfile)
				description := generateBuildDescription(entry, dockerfile)
				containerName := generateContainerName(entry, namingLookup, dirname)
				fmt.Println(command + " " + containerName)
				newResult := parsers.BuildResult{BuildArgs: entry, CommandPrefix: command,
					Description: description, Filename: dockerfile, Parser: "dockerbuild",
					Name: subpath, ContainerName: containerName}
				results = append(results, newResult)
			}
		}

		// Parse into json
		outJson, _ := json.Marshal(results)
		output := string(outJson)

		// If it's empty, still provide an empty list
		if output == "" {
			output = "[]"
		}

		// If we are running in a GitHub Action, set the outputs
		if utils.IsGitHubAction() {
			fmt.Printf("::set-output name=dockerbuild_matrix::%s\n", output)
		}
	}
	return nil
}

// Create a build matrix from an existing specification (we trust that it is correct)
func NewBuildMatrix(matrixArgs map[string][]string) []map[string]string {

	// The final result is a list of key value pairs
	results := []map[string]string{}

	// First get the min length to loop through
	minLength := 100
	for _, values := range matrixArgs {
		if len(values) < minLength {
			minLength = len(values)
		}
	}

	// Now go through the min length of each
	count := 1
	for key, values := range matrixArgs {

		// If we are in the first loop, create the original list of results
		if count == 1 {
			for _, value := range values {
				entry := make(map[string]string)
				entry[key] = value
				results = append(results, entry)
			}
		} else {
			for i, value := range values {

				// If we've exceeded the min length, we don't have a perfect match
				if i+1 > len(results) {
					break
				}
				results[i][key] = value
			}
		}
		count += 1
	}
	return results
}

// GetBuildMatrix generates a build matrix, across all variable options
func GetBuildMatrix(vars []parsers.BuildVariable) []map[string]string {

	// The final result is a list of key value pairs
	results := []map[string]string{}

	for _, buildvar := range vars {
		results = getBuildMatrix(buildvar.Name, buildvar.Values, results)
	}
	return results
}

// generateContainerName creates a suggested name for the container (without registry)
func generateContainerName(buildargs map[string]string, lookup map[string][]ContainerNamer, basename string) string {

	// Start with the container basename (usually the directory it is in)
	containerName := basename

	// For each known container variable, this gets added to the container name
	for _, namer := range lookup["container"] {
		containerName = containerName + "-" + namer.Slug + "-" + buildargs[namer.Key]
	}

	// Add tags, if there are any
	if len(lookup["tag"]) > 0 {
		containerName += ":"
		for i, namer := range lookup["tag"] {
			containerName = containerName + namer.Slug + "-" + buildargs[namer.Key]
			if i != len(lookup["tag"])-1 {
				containerName = containerName + "-"
			}
		}
		containerName = strings.Trim(containerName, "-")
	}
	return containerName
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
			newEntry := make(map[string]string)

			// This copies the previous entry
			for k, v := range entry {
				newEntry[k] = v
			}
			newEntry[newkey] = value
			updated = append(updated, newEntry)
		}
	}
	return updated
}
