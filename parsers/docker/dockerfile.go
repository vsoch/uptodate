package docker

// The Dockerfile parser is optimized to find and update FROM statements

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"

	df "github.com/asottile/dockerfile"
	"github.com/vsoch/uptodate/parsers"
	"github.com/vsoch/uptodate/parsers/git"
	"github.com/vsoch/uptodate/utils"
	"path/filepath"
)

// Command extends a command to perform custom parsing functions
// https://github.com/asottile/dockerfile/blob/master/parse.go#L14
type Command struct {
	Cmd       string   // lowercased command name (ex: `from`)
	SubCmd    string   // for ONBUILD only this holds the sub-command
	Json      bool     // whether the value is written in json form
	Original  string   // The original source line
	StartLine int      // The original source line number which starts this command
	EndLine   int      // The original source line number which ends this command
	Flags     []string // Any flags such as `--from=...` for `COPY`.
	Value     []string // The contents of the command (ex: `ubuntu:xenial`)
}

// An update to a FROM includes the original content and update
type Update struct {
	Original string
	Updated  string
	LineNo   int
}

// StartIndex is the StartLine -1 (for indexing)
func (c *Command) StartIndex() int {
	return c.StartLine - 1
}

// EndIndex is the EndLine -1 (for indexing)
func (c *Command) EndIndex() int {
	return c.EndLine - 1
}

// Dockerfile holds commands, path, and raw Dockerfile content
type Dockerfile struct {
	Root    string
	Path    string
	Raw     string
	Cmds    map[string][]Command // Lookup by command type for quicker parsing
	Updates []Update
}

// Determine if a Dockerfile contains BUILD args
func (d *Dockerfile) HasBuildArgs() bool {

	// If we don't have commands yet, try to parse
	if len(d.Cmds) == 0 {
		d.ParseCommands()
	}

	if values, ok := d.Cmds["arg"]; ok {
		return len(values) > 0
	}

	// We don't have any build args
	return false
}

// Return the basename of the Dockerfile
func (d *Dockerfile) BaseName() string {
	return filepath.Base(d.Path)
}

// Return the relative path of the dockerfile, which we should strip for GHA
func (d *Dockerfile) RelativePath() string {

	relpath, err := filepath.Rel(d.Root, d.Path)
	if err != nil {
		log.Fatal(err)
	}
	return relpath

}

// AddCommands parses a set of df.Commands into uptodate Commands
func (d *Dockerfile) AddCommands(cmds []df.Command) {

	// Init map of commands
	d.Cmds = make(map[string][]Command)

	// Add each command
	for _, cmd := range cmds {
		d.AddCommand(cmd)
	}
}

// AddCommand parses a df.Command into a uptodate Command and adds to list
func (d *Dockerfile) AddCommand(cmd df.Command) {
	extendedCmd := Command(cmd)

	// We only care about FROM statements, but maybe others in the future
	commandType := strings.ToLower(cmd.Cmd)
	if utils.IncludesString(commandType, []string{"from", "arg"}) {

		// Add to lookup, checking if key already exists
		if _, ok := d.Cmds[commandType]; ok {
			d.Cmds[commandType] = append(d.Cmds[commandType], extendedCmd)
		} else {
			d.Cmds[commandType] = []Command{extendedCmd}
		}
	}
}

// UpdateFroms, meaning we look for newer versions, etc.
func (d *Dockerfile) UpdateFroms() {

	// Prepare a set of updates
	d.Updates = []Update{}

	// Loop through FROMs and update!
	for _, from := range d.Cmds["from"] {

		// An "empty" update will be returned if nothing to do
		newUpdate := UpdateFrom(from.Value)
		if !reflect.DeepEqual(newUpdate, Update{}) {

			newUpdate.Original = from.Original
			newUpdate.LineNo = from.StartIndex()
			d.Updates = append(d.Updates, newUpdate)
		}

	}
}

// ReplaceFroms simply replaces found FROM with a known value
// This is typically run instead of UpdateFroms
func (d *Dockerfile) ReplaceFroms(name string, tag string) {

	// Prepare a set of updates
	d.Updates = []Update{}

	// Loop through FROMs and update! See UpdateFroms for comments
	for _, from := range d.Cmds["from"] {

		container := from.Value[0]

		// We can't reliably replace a variable
		isVariable := strings.Contains(container, "$")
		if isVariable {
			continue
		}

		// Just get rid of hashes and tags so we have base container name
		if strings.Contains(container, "@") {
			parts := strings.SplitN(container, "@", 2)
			container = parts[0]
		}

		if strings.Contains(container, ":") {
			parts := strings.SplitN(container, ":", 2)
			container = parts[0]
		}

		// Clean up white spaces, and check if we have a match
		container = strings.Trim(container, " ")
		if container == name {

			updated := container + ":" + tag

			// Add original content back
			for _, extra := range from.Value[1:] {
				updated += " " + extra
			}

			update := Update{Original: from.Original, Updated: updated, LineNo: from.StartIndex()}
			d.Updates = append(d.Updates, update)
		}

	}
}

// Write writes a new Dockerfile
func (d *Dockerfile) Write() {

	// Read in the raw Dockerfile
	raw := utils.ReadFile(d.Path)

	// Split into original lines
	lines := strings.Split(raw, "\n")

	// For each Update, replace exact line with new version
	for _, update := range d.Updates {
		fmt.Printf("Updating %s to %s\n", update.Original, update.Updated)

		// This ensures we keep the tag preserved for future checks, but change the file so it rebuilds
		lines[update.LineNo] = "FROM " + update.Updated
	}
	content := strings.Join(lines, "\n")
	utils.WriteFile(d.Path, content)
}

// DockerfileParser holds one or more Dockerfile
type DockerfileParser struct {
	Dockerfiles []Dockerfile
}

// AddDockerfile adds a Dockerfile to the Parser
// Not super efficient, but reasonably we don't have that many Dockerfile
func (s *DockerfileParser) CountUpdated() int {
	count := 0
	for _, dockerfile := range s.Dockerfiles {
		count += len(dockerfile.Updates)
	}
	return count
}

// ParseCommands in a Dockerfile, reading the file and saving a subest
func (d *Dockerfile) ParseCommands() {

	cmds, err := df.ParseFile(d.Path)

	// If we can't read for whatever reason, log the issue and continue
	if err != nil {
		log.Printf("%s is not a loadable Dockerfile, skipping.", d.Path)
		return
	}
	// Add commands, parse FROMs, and LABELS
	d.AddCommands(cmds)
}

// AddDockerfile adds a Dockerfile to the Parser
func (s *DockerfileParser) AddDockerfile(root string, path string) {

	// Create a new Dockerfile entry
	dockerfile := Dockerfile{Path: path, Root: root}
	dockerfile.ParseCommands()
	dockerfile.UpdateFroms()
	s.Dockerfiles = append(s.Dockerfiles, dockerfile)
}

// Entrypoint to parse one or more Dockerfiles
func (s *DockerfileParser) Parse(path string, dryrun bool, changesOnly bool) error {

	// Find Dockerfiles in path and allow prefixes
	paths, _ := utils.RecursiveFind(path, "Dockerfile", true)

	// If we want changed only, honor that
	if changesOnly {

		// Create list of changes (Modify or Add)
		changed := git.GetChangedFilesStrings(path)
		paths = utils.FindOverlap(paths, changed)
	}

	// No updated?
	if len(paths) == 0 {
		fmt.Println("No changes to parse.")
	}

	// Add each path as a Dockerfile to the parser to update
	for _, subpath := range paths {
		s.AddDockerfile(path, subpath)
	}

	// Keep track of updated count and set of results
	count := 0
	results := []parsers.Result{}

	// Do we have updates? Count and write to file
	if s.CountUpdated() > 0 {
		for _, dockerfile := range s.Dockerfiles {
			if len(dockerfile.Updates) == 0 {
				continue
			}

			// Only write changes if it's not a dryrun
			if !dryrun {

				// Add a new result to print later
				result := parsers.Result{Filename: dockerfile.Path, Name: dockerfile.Path, Parser: "dockerfile"}
				results = append(results, result)
				dockerfile.Write()
			}
			count += 1
		}

	}
	action := "Updated"
	if dryrun {
		action = "Will Be Updated"
	}
	fmt.Println("\n  ⭐️ " + action + " ⭐️")
	fmt.Printf("     Checked: %d\n", len(s.Dockerfiles))
	if !dryrun {
		fmt.Printf("    Modified: %d\n", count)
	}

	// If we are running in a GitHub Action, set the outputs
	if utils.IsGitHubAction() {
		outJson, _ := json.Marshal(results)
		output := string(outJson)
		if output == "" {
			output = "[]"
		}
		fmt.Printf("::set-output name=dockerfile_matrix::%s\n", output)
	}
	return nil
}
