package docker

// The Dockerfile parser is optimized to find and update FROM statements

import (
	/*	"encoding/json"*/
	"fmt"
	"github.com/vsoch/uptodate/config"
	"github.com/vsoch/uptodate/parsers"
	"github.com/vsoch/uptodate/utils"
	"log"
	"reflect"
)

// BuildArg is a key with one or more values (can be versions)
type BuildArg struct {

	// Versions are optional, and only valid for manual
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

// BuildArgSpack expects metadata for building from spack
type BuildArgSpack struct {
	Type     string   `json:"type"`
	Name     string   `json:"name,omitempty"`
	StartAt  string   `json:"startat,omitempty"`
	Versions []string `json:"versions,omitempty"`
	Filter   []string `json:"filter,omitempty"`
	Skips    []string `json:"skips,omitempty"`
	AsView   bool     `json:"view,omitempty"`
}

// BuildArgContainer updates container tags, etc.
type BuildArgContainer struct {
	Type     string   `json:"type"`
	Name     string   `json:"name,omitempty"`
	StartAt  string   `json:"startat,omitempty"`
	Versions []string `json:"versions,omitempty"`
	Filter   []string `json:"filter,omitempty"`
	Skips    []string `json:"skips,omitempty"`
}

// DockerBuild holds one or more build args
type DockerBuild struct {
	BuildArgs []map[string]interface{}
}

// DockerBuildParser holds one or more Docker Builds
type DockerBuildParser struct {
	Builds []DockerBuild
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
			log.Fatal("dockerbuild section not detected in config!")
		}

		// Prepare a matrix of json results
		results := parsers.BuildResult{}

		for key, buildarg := range conf.DockerBuild.BuildArgs {

			// If it has a type, it either is that type, or we map to another type
			// TODO each of these needs to be parsed, then output matrix!
			if buildarg.Type == "container" {
				fmt.Println("Found container build arg", key, buildarg)
			} else if buildarg.Type == "spack" {
				fmt.Println("Found spack build arg", key, buildarg)
			} else {
				fmt.Println("Found regular build arg", key, buildarg)
			}
		}
		fmt.Println(results)
	}
	return nil
}
