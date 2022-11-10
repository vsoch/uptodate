package docker

// The Docker bases parser is optimized to find and update FROM statements

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
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
type DockerBasesParser struct{}

// Entrypoint to parse one or more Docker build matrices
func (s *DockerBasesParser) Parse(basesPath string, providedPaths []string, changesOnly bool, branch string, registry string, buildAll bool) error {

	// Find bases (Dockerfile under path)
	bases, _ := utils.RecursiveFind(basesPath, "Dockerfile", true)
	if len(bases) == 0 {
		fmt.Printf("No Dockerfile found under %s, nothing to build.", basesPath)
		return nil
	}

	// Parse each path
	paths := []string{}
	for _, path := range providedPaths {

		// Find Dockerfiles in path and allow prefixes
		newPaths, _ := utils.RecursiveFind(path, "uptodate.yaml", true)

		// If we want changed only, honor that, unless registry is defined
		if changesOnly && registry == "" {

			// Create list of changes (Modify or Add)
			changed := git.GetChangedFilesStrings(path, branch)
			newPaths = utils.FindOverlap(newPaths, changed)
		}
		paths = append(paths, newPaths...)
	}

	// No updated?
	if len(paths) == 0 {
		fmt.Println("No changes to parse.")
	}

	// Prepare a list of build results
	results := []parsers.BuildResult{}

	// Look at each found path, parse into build matrix
	for _, subpath := range paths {
		conf := config.Load(subpath)

		// We must have a DockerBuild to continue! This checks against an empty one
		if reflect.DeepEqual(conf.DockerBuild, config.DockerBuild{}) {
			log.Printf("dockerbuild section not detected in config, skipping %s\n", subpath)
			continue
		}

		// If the builf isn't active, skip
		if !conf.DockerBuild.Active {
			fmt.Printf("Skipping %s, not active\n", subpath)
			continue
		}

		// Get a matrix, either from the config or on the fly generation, and naming lookup
		namingLookup := make(map[string][]ContainerNamer)
		namingList := []ContainerNamer{}
		matrix := GetBuildMatrix(conf, &namingLookup, &namingList, &conf.DockerBuild.Exclude)

		// Find Dockerfile in subpath
		dirnamePath := filepath.Dir(subpath)
		dirname := filepath.Base(dirnamePath)

		// For each container name, look up latest variables and generate labels lookup
		latestValues := getLatestValues(registry, matrix, namingLookup, namingList, dirname, conf.DockerBuild.ContainerBasename)

		// Now get current values for each container (e.g., hashes)
		currentValues := getCurrentValues(registry, matrix, namingLookup, dirname, conf.DockerBuild.ContainerBasename)

		// We need a new build for each Dockerfile bases found (hopefully not many)
		for _, dockerfile := range bases {
			for _, entry := range matrix {

				// Suffix to the container is the location of the base image
				relDirname := dirname
				relpath := strings.Trim(strings.ReplaceAll(filepath.Dir(dockerfile), basesPath, ""), string(os.PathSeparator))
				if relpath != "" {
					relpath = strings.Trim(strings.ReplaceAll(relpath, string(os.PathSeparator), "-"), "-")
					relDirname = relDirname + "-" + relpath
				}

				// We can look up variables in the config
				containerName := generateContainerName(registry, entry, namingLookup, relDirname, conf.DockerBuild.ContainerBasename)

				// Get labels for the container
				labels := getLabelLookup(entry, namingLookup, latestValues)

				// Should we include for the build?
				includeContainer := compareWithLatest(containerName, latestValues, currentValues, buildAll)
				if includeContainer {
					command := generateBasesBuildCommand(entry, dockerfile, labels)
					description := generateBuildDescription(entry, dockerfile)
					fmt.Println(command + " " + containerName)

					// Generate a build for each Dockerfile in bases path
					newResult := parsers.BuildResult{BuildArgs: entry, CommandPrefix: command,
						Description: description, Filename: dockerfile, Parser: "dockerbases",
						Name: subpath, ContainerName: containerName, Context: dirnamePath}

					results = append(results, newResult)

				}
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
		utils.WriteGitHubOutput("dockerbases_matrix_empty", strconv.FormatBool(isEmpty))
		utils.WriteGitHubOutput("dockerbases_matrix", output)
	}
	return nil
}

// generateBasesCommand will generate a build command for a given bases Dockerfile and buildards
func generateBasesBuildCommand(buildargs map[string]string, dockerfile string, labels map[string]string) string {

	// Start the command (use environment variable for name)
	command := "docker build -f " + dockerfile

	// Add each buildarg and labels
	for key, value := range buildargs {
		command += " --build-arg " + key + "=" + value
	}
	for key, value := range labels {
		command += " --label " + key + "=" + value
	}
	return command
}
