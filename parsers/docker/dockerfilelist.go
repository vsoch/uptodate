package docker

// The Dockerfile parser is optimized to find and update FROM statements

import (
	"encoding/json"
	"fmt"

	"github.com/vsoch/uptodate/parsers"
	"github.com/vsoch/uptodate/utils"
)

// DockerfileListParser holds one or more Dockerfile
type DockerfileListParser struct {
	Dockerfiles []Dockerfile
}

// AddDockerfile adds a Dockerfile to the Parser
func (s *DockerfileListParser) AddDockerfile(root string, path string) {

	// Create a new Dockerfile entry
	dockerfile := Dockerfile{Path: path, Root: root}
	s.Dockerfiles = append(s.Dockerfiles, dockerfile)
}

// Entrypoint to parse one or more Dockerfiles
func (s *DockerfileListParser) Parse(path string) error {

	// Find Dockerfiles in path and allow prefixes
	paths, _ := utils.RecursiveFind(path, "Dockerfile", true)

	// Add each path as a Dockerfile to the parser to update
	for _, subpath := range paths {
		s.AddDockerfile(path, subpath)
	}

	// Keep track of updated count and set of results
	results := []parsers.Result{}

	// Print each dockerfile to the console
	for _, dockerfile := range s.Dockerfiles {

		// Add a new result to print later
		result := parsers.Result{Filename: dockerfile.Path, Name: dockerfile.Path, Parser: "dockerfilelist"}
		results = append(results, result)
		fmt.Println(dockerfile.Path)

	}

	// If we are running in a GitHub Action, set the outputs
	if utils.IsGitHubAction() {
		outJson, _ := json.Marshal(results)
		fmt.Printf("::set-output name=dockerfilelist_matrix::%s\n", string(outJson))
	}
	return nil
}
