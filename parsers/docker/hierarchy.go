package docker

// The Dockerfile hierarchy parser expects a directory with subfolders
// organized by tags, with a top level config yaml to indicate preferences
// for parsing the hierarchy. Each subfolder should have a Dockerfile:
//
// ubuntu/
//    uptodate.yaml
//    latest/
//      Dockerfile
//    20.04/
//      Dockerfile
//    18.04/
//      Dockerfile
//
// By default, the top level folder is identified by presence of an uptodate.yaml.
// This file holds preferences for filtering (tag discovery) and the image name
// and eventually other preferences/metadata that will be nice to have.

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	fpath "path"
	"strconv"

	df "github.com/asottile/dockerfile"
	"github.com/vsoch/uptodate/config"
	"github.com/vsoch/uptodate/parsers"
	"github.com/vsoch/uptodate/utils"
	"path/filepath"
	"reflect"
)

// A DockerHierarchy holds a set of preferences for parsing docker hierarchies
type DockerHierarchy struct {
	Root            string
	Path            string // path to uptodate.yaml
	Container       string
	Filters         []string
	StartAtVersion  string
	EndAtVersion    string
	SkipVersions    []string
	IncludeVersions []string
	tags            []string
}

// Return the basename of the specific root
func (d *DockerHierarchy) BaseName() string {
	return filepath.Base(d.Path)
}

// Return the directory name of the root
func (d *DockerHierarchy) DirName() string {
	return filepath.Dir(d.Path)
}

// GetLatestDockerfile of existing container within user preferences
func (d *DockerHierarchy) GetLatestDockerfile(tag string) string {

	// The Dockerfile should be under <root>/<tag>/Dockerfile
	dockerfile := fpath.Join(d.DirName(), tag, "Dockerfile")
	if _, err := os.Stat(dockerfile); os.IsNotExist(err) {
		log.Fatalf("%s does not exist.\n", dockerfile)
	}
	return dockerfile
}

// CopyDockerfile will copy a template into a directory, creating if doesn't exist
// This assumes the directory is a child of the root
func (d *DockerHierarchy) CopyDockerfile(template string, tag string) string {

	// The dirname must be a child of the parent docker hierarchy
	destDir := fpath.Join(d.DirName(), tag)
	dest := fpath.Join(destDir, "Dockerfile")

	// If the Output directory doesn't exist, create it
	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		err := os.Mkdir(destDir, 0755)
		if err != nil {
			log.Fatal("Cannot create", destDir)
		}
	}

	// Copy Dockerfile there, return new path
	fmt.Printf("Copying %s to %s\n", template, dest)
	utils.CopyFile(template, dest)
	return dest
}

// Read in an existing Dockerfile and update top FROM version
func (d *DockerHierarchy) UpdateFrom(path string, tag string) {

	// Create a new Dockerfile entry
	dockerfile := Dockerfile{Path: path}
	cmds, err := df.ParseFile(path)

	// If we can't read for whatever reason, log the issue and continue
	if err != nil {
		log.Printf("%s is not a loadable Dockerfile, skipping.", path)
		return
	}

	// Add commands, parse FROMs, and then update matching FROMs to new tag
	dockerfile.AddCommands(cmds)
	dockerfile.ReplaceFroms(d.Container, tag)

}

// Dockerfile holds commands, path, and raw Dockerfile content
type DockerHierarchyParser struct {
	Path string

	// Roots that each contain a DockerHierarchy
	Roots []DockerHierarchy
}

// Return the basename of the Hierarchy
func (d *DockerHierarchyParser) BaseName() string {
	return filepath.Base(d.Path)
}

// Entrypoint to parse a docker hierarchy directory
func (s *DockerHierarchyParser) Parse(path string, dryrun bool) {

	// TODO allow the user to specify args instead of auto-discovery
	s.Load(path)

	// Run the updater
	s.Update(dryrun)
}

// Load will parse the configs for one or more docker hierarchy directories
func (s *DockerHierarchyParser) Load(path string) {

	// Find config files in path and don't allow prefixes
	paths, _ := utils.RecursiveFind(path, "uptodate.yaml", false)

	// Look at each found path
	for _, subpath := range paths {
		conf := config.Load(subpath)

		// If the dockerhierarchy key is missing, we cannot parse!
		var emptyDockerHierarchy config.DockerHierarchy
		if reflect.DeepEqual(conf.DockerHierarchy, emptyDockerHierarchy) {
			fmt.Printf("dockerhierarchy key is missing from %s, skipping.\n", subpath)
			continue
		}

		// If we don't have filters, add a standard that looks for version
		if len(conf.DockerHierarchy.Container.Filter) == 0 {
			conf.DockerHierarchy.Container.Filter = append(conf.DockerHierarchy.Container.Filter, parsers.VersionRegex)
		}

		// Create a new DockerHierarchy, set name and filters
		hier := DockerHierarchy{Container: conf.DockerHierarchy.Container.Name,
			Filters:         conf.DockerHierarchy.Container.Filter,
			StartAtVersion:  conf.DockerHierarchy.Container.StartAt,
			EndAtVersion:    conf.DockerHierarchy.Container.EndAt,
			SkipVersions:    conf.DockerHierarchy.Container.Skips,
			IncludeVersions: conf.DockerHierarchy.Container.Includes,
			Path:            subpath,
			Root:            path}

		// Add the hierarchy to those we know about
		s.Roots = append(s.Roots, hier)
	}
}

// Update will look at existing tags, and compare to known and write new files
func (s *DockerHierarchyParser) Update(dryrun bool) error {

	// Save all results for later use
	results := []parsers.Result{}

	// For each root, derive updates!
	for _, root := range s.Roots {

		// Get all versions (tags) based on filters and user preferences
		versions := GetVersions(root.Container, root.Filters, root.StartAtVersion, root.EndAtVersion, root.SkipVersions, root.IncludeVersions)

		// At this point we have a list of versions we want.
		// We now compare existing to those that need to be created
		containerDir := root.DirName()

		// list directory, include dirs, not files
		subDirs := utils.ListDir(containerDir, true, false)

		// Find the difference in the lists - what tags we have
		// that are not present on the filesystem
		// Return strings that are in the first list but not the second
		missing := utils.FindMissingInSecond(versions, subDirs)
		present := utils.FindOverlap(subDirs, versions)

		// If dry run, just print to screen
		if dryrun {
			fmt.Println("\n  ⭐️ Will Be Updated ⭐️")
			fmt.Printf("     Missing versions for %s: %s\n", root.Container, missing)
			fmt.Printf("     Present versions for %s: %s\n", root.Container, present)
			return nil
		}

		// If we have versions to create and no templates, no go
		if len(missing) > 0 && len(present) == 0 {
			log.Fatal("There are missing tags but no existing Dockerfile present to copy, cannot continue.")
		}

		if len(present) == 0 {
			fmt.Println("No Dockerfile found to parse.")
			continue
		}

		// Get the latest Dockerfile that exists
		dockerfile := root.GetLatestDockerfile(present[len(present)-1])

		// For each new container to create, copy the previous Dockerfile, update verison
		for _, miss := range missing {
			newDockerfile := root.CopyDockerfile(dockerfile, miss)
			root.UpdateFrom(newDockerfile, miss)

			// Add the result as updated to the list
			result := parsers.Result{Filename: newDockerfile, Identifier: miss, Name: newDockerfile, Parser: "dockerhierarchy"}
			results = append(results, result)

		}

		// Update stats
		fmt.Println("\n  ⭐️ Updated ⭐️")
		fmt.Printf("     Updated versions for %s: %s\n", root.Container, missing)
		fmt.Printf("     Present versions for %s: %s\n", root.Container, present)

	}

	// If we are running in a GitHub Action, set the outputs
	if utils.IsGitHubAction() {
		outJson, _ := json.Marshal(results)
		output := string(outJson)
		isEmpty := false
		if output == "" {
			output = "[]"
			isEmpty = true
		}
		fmt.Printf("::set-output name=dockerhierarchy_matrix::%s\n", output)
		fmt.Printf("::set-output name=dockerhierarchy_matrix_empty::%s\n", strconv.FormatBool(isEmpty))
	}

	return nil
}
