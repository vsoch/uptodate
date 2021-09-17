package docker

// The Dockerfile parser is optimized to find and update FROM statements

import (
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/vsoch/uptodate/config"
	"github.com/vsoch/uptodate/parsers"
	"github.com/vsoch/uptodate/parsers/git"
	"github.com/vsoch/uptodate/utils"
)

// DockerBuildParser holds one or more Docker Builds
type DockerBuildParser struct{}

// Keep track of an original lookup key and the slug for the buildarg
type ContainerNamer struct {
	Slug string
	Key  string
	Type string
}

type Label struct {
	Original string
	Type     string
	Name     string
	Value    string
}

// Entrypoint to parse one or more Docker build matrices
func (s *DockerBuildParser) Parse(path string, changesOnly bool, branch string, registry string) error {

	// Find config files in path and don't allow prefixes
	paths, _ := utils.RecursiveFind(path, "uptodate.yaml", false)

	// If we want changed only, honor that, unless registry is defined
	if changesOnly && registry == "" {

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

		// Get a matrix, either from the config or on the fly generation, and naming lookup
		namingLookup := make(map[string][]ContainerNamer)
		namingList := []ContainerNamer{}
		matrix := GetBuildMatrix(conf, &namingLookup, &namingList)

		// Prepare a list of build results
		results := []parsers.BuildResult{}

		// Find Dockerfile in subpath
		dirnamePath := filepath.Dir(subpath)
		dockerfiles, _ := utils.RecursiveFind(dirnamePath, "Dockerfile", true)
		dirname := filepath.Base(dirnamePath)

		// For each container name, look up latest variables and generate labels lookup
		latestValues := getLatestValues(matrix, namingList)

		// Now get current values for each container (e.g., hashes)
		currentValues := getCurrentValues(registry, matrix, namingLookup, latestValues, dirname)

		// We need a new build for each Dockerfile found (hopefully not many)
		for _, dockerfile := range dockerfiles {
			for _, entry := range matrix {

				// We can look up variables in the config
				containerName := generateContainerName(registry, entry, namingLookup, dirname)

				// Get labels for the container
				labels := getLabelLookup(entry, namingLookup, latestValues)

				// Should we include for the build?
				includeContainer := compareWithLatest(containerName, latestValues, currentValues)
				if includeContainer {
					command := generateBuildCommand(entry, dockerfile, labels)
					description := generateBuildDescription(entry, dockerfile)
					fmt.Println(command + " " + containerName)
					newResult := parsers.BuildResult{BuildArgs: entry, CommandPrefix: command,
						Description: description, Filename: dockerfile, Parser: "dockerbuild",
						Name: subpath, ContainerName: containerName}
					results = append(results, newResult)
				}
			}
		}

		// Parse into json
		outJson, _ := json.Marshal(results)
		output := string(outJson)

		// If it's empty, still provide an empty list
		isEmpty := false
		if output == "" {
			output = "[]"
			isEmpty = true
		}

		// If we are running in a GitHub Action, set the outputs
		if utils.IsGitHubAction() {
			fmt.Printf("::set-output name=dockerbuild_matrix_empty::%s\n", strconv.FormatBool(isEmpty))
			fmt.Printf("::set-output name=dockerbuild_matrix::%s\n", output)
		}
	}
	return nil
}

// compareWithLatest compares current with latest values, and (assuming we have all variables that exist)
// determines if we should run the build.
func compareWithLatest(containerName string, latestValues map[string]string, currentValues map[string]map[string]Label) bool {

	// By default be conservative and rebuild
	currentLabels, ok := currentValues[containerName]
	if !ok {
		return true
	}

	// Break apart latest values into keys and tags
	var lookup = map[string]map[string]string{}
	for latest, value := range latestValues {
		parts := strings.SplitN(latest, ":", 2)
		lookup[parts[0]] = map[string]string{}

		// lookup[<keyM][<tag>] = value
		lookup[parts[0]][parts[1]] = value
	}

	// For each current label, if we have a matching, check against
	for _, label := range currentLabels {
		fmt.Println(label)
	}
	return true
}

// getCurrentValues retrieves and parses current container label values
func getCurrentValues(registry string, matrix []map[string]string, namingLookup map[string][]ContainerNamer,
	latest map[string]string, dirname string) map[string]map[string]Label {

	// Prepare current values
	var currentValues = map[string]map[string]Label{}

	// No registry, no ability to check anything because we only can check containers
	if registry == "" {
		return currentValues
	}

	// For each container name, we use the name as a lookup
	for _, entry := range matrix {

		// We can look up variables in the config
		containerName := generateContainerName(registry, entry, namingLookup, dirname)
		fmt.Println(containerName)

		withoutTag := strings.SplitN(containerName, ":", 2)[0]

		// Get a list of known tags to start
		tags := GetImageTags(withoutTag)

		if len(tags) == 9 {
			fmt.Printf("Container %s does not have any tags, skipping lookup.", withoutTag)
			continue
		}

		// Prepare an entry for the container name
		currentValues[containerName] = map[string]Label{}
		imageConf := GetImageConfig(containerName)

		// For each label in the image conf, if it matches an uptodate_matrix, save it!
		for _, original := range imageConf.Config.Labels {
			if strings.HasPrefix(original, "uptodate_matrix_") {

				// Should split into uptodate_matrix_<type>_<key>=<value>
				label := strings.Replace(original, "uptodate_matrix_", "", 1)

				// <type>_<key>=<value>
				parts := strings.SplitN(label, "=", 2)
				value := parts[1]

				// <type>_<key>
				parts = strings.SplitN(parts[0], "_", 2)
				argType := parts[0]
				argKey := parts[1]
				currentValues[containerName][argKey] = Label{Original: original, Name: argKey, Type: argType, Value: value}
			}
		}
	}
	return currentValues
}

// getLatestValues returns a lookup of latest build arg namers and tags tha
func getLatestValues(matrix []map[string]string, namingList []ContainerNamer) map[string]string {

	// current values for different build args
	var currentValues = map[string]string{}

	for _, entry := range matrix {

		// Get updated values for each known build container argument
		for _, namer := range namingList {

			// We can only look for updated hashes for containers
			if namer.Type == "container" {

				// Do we have a current tag?
				tag, ok := entry[namer.Key]
				if !ok {
					continue
				}

				// If we already have an entry, continue
				if _, ok := currentValues[namer.Key+":"+tag]; ok {
					continue
				}

				currentValues[namer.Key+":"+tag] = getUpdatedContainer(namer.Slug + ":" + tag)
			}
		}

	}
	return currentValues
}

// GetBuildMatrix: Upper level function to get a build matrix, either from config or generation
func GetBuildMatrix(conf config.Conf, namingLookup *map[string][]ContainerNamer, namingList *[]ContainerNamer) []map[string]string {

	// Prepare naming lookup
	(*namingLookup)["container"] = []ContainerNamer{}
	(*namingLookup)["tag"] = []ContainerNamer{}

	// Keys we will skip if not included in matrix
	allowKeys := []string{}

	// Prepare lists of values to create a matrix over
	vars := []parsers.BuildVariable{}

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
			namer.Type = "container"
			(*namingLookup)["container"] = append((*namingLookup)["container"], namer)
			(*namingList) = append((*namingList), namer)
		} else if buildarg.Type == "spack" {
			result := parseSpackBuildArg(key, buildarg)
			vars = append(vars, result...)
			namer.Type = "spack"
			(*namingLookup)["tag"] = append((*namingLookup)["tag"], namer)
			(*namingList) = append((*namingList), namer)
		} else {
			result := parseBuildArg(key, buildarg)
			vars = append(vars, result...)
			namer.Type = "manual"
			(*namingLookup)["tag"] = append((*namingLookup)["tag"], namer)
			(*namingList) = append((*namingList), namer)
		}
	}

	// If we don't have the matrix yet, create all possible combinations
	if len(conf.DockerBuild.Matrix) == 0 {
		matrix = GenerateBuildMatrix(vars)
	}
	return matrix
}

// Create a NEW build matrix from an existing specification (we trust that it is correct)
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

// GenerateBuildMatrix generates a build matrix, across all variable options
func GenerateBuildMatrix(vars []parsers.BuildVariable) []map[string]string {

	// The final result is a list of key value pairs
	results := []map[string]string{}
	for _, buildvar := range vars {
		results = generateBuildMatrix(buildvar.Name, buildvar.Values, results)
	}
	return results
}

// generateContainerName creates a suggested name for the container (without registry)
func generateContainerName(registry string, buildargs map[string]string, lookup map[string][]ContainerNamer, basename string) string {

	// If given a registry name, use it
	if registry != "" {
		basename = registry + "/" + basename
	}

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
func generateBuildCommand(buildargs map[string]string, dockerfile string, labels map[string]string) string {

	// The build should be relative to where the Dockerfile is
	filename := filepath.Base(dockerfile)

	// Start the command (use environment variable for name)
	command := "docker build -f " + filename

	// Add each buildarg and labels
	for key, value := range buildargs {
		command += " --build-arg " + key + "=" + value
	}
	for key, value := range labels {
		command += " --label " + key + "=" + value
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

// getLabelLookup for each container
func getLabelLookup(buildargs map[string]string, lookup map[string][]ContainerNamer, latestValues map[string]string) map[string]string {

	var labels = map[string]string{}

	// We can only currently generate labels (and update) containers
	for _, namer := range lookup["container"] {
		key := namer.Key + ":" + buildargs[namer.Key]
		value, ok := latestValues[key]
		if ok {
			labels["uptodate_matrix_"+namer.Type+"_"+namer.Key] = value
		}
	}
	return labels
}

// generateBuildMatrix is a helper function to grow a list of maps with each set of params
func generateBuildMatrix(newkey string, values []string, previous []map[string]string) []map[string]string {

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
