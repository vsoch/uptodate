package parsers

// The Dockerfile parser is optimized to find and update FROM statements

import (
	"fmt"
	"log"
	"strings"

	lookout "github.com/alecbcs/lookout/update"
	df "github.com/asottile/dockerfile"
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
	Path    string
	Raw     string
	Cmds    map[string][]Command // Lookup by command type for quicker parsing
	Updates []Update
}

// Return the basename of the Dockerfile
func (d *Dockerfile) BaseName() string {
	return filepath.Base(d.Path)
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
	if utils.IncludesString(commandType, []string{"from"}) {

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

		// This is the full container name, e.g., ubuntu:16.04
		container := from.Value[0]

		// Keep the original for later comparison
		original := strings.Join(from.Value, " ")

		// Variable statements we can't reliably update
		isVariable := strings.Contains(container, "$")
		if isVariable {
			continue
		}

		// We want to keep track of having a hash and/or tag
		hasHash := false
		hasTag := false

		// First remove any digest from the container
		if strings.Contains(container, "@") {
			parts := strings.SplitN(container, "@", 2)
			container = parts[0]
			hasHash = true
		}

		// Now extract any tag from the container
		tag := "latest"
		if strings.Contains(container, ":") {
			parts := strings.SplitN(container, ":", 2)
			container = parts[0]
			tag = parts[1]
			hasTag = true
		} else {
			fmt.Printf("No tag specified for %s, will default to latest.\n", container)
		}

		// If it has a hash but no digest, we can't correctly parse
		if hasHash && !hasTag {
			fmt.Printf("Cannot parse %s, has a hash but no tag, cannot be looked up.\n", container)
			continue
		}

		// Get the updated container hash for the tag
		url := container + ":" + tag
		out, found := lookout.CheckUpdate("docker://" + url)

		if found {
			// Prepare the updated string, the result.Name is digest
			result := *out
			updated := url + "@" + result.Name

			// Add original content back
			for _, extra := range from.Value[1:] {
				updated += " " + extra
			}

			// If the updated version is different from the original, update
			if updated != original {

				// TODO I've never seen a multi-line FROM, but this will need
				// adjustment if one exists to replace a range of lines
				update := Update{Original: from.Original, Updated: updated, LineNo: from.StartIndex()}
				d.Updates = append(d.Updates, update)
			}

		} else {
			fmt.Println("Cannot find container URI", url)
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
		fmt.Printf("\nUpdating %s to %s\n", update.Original, update.Updated)

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

// AddDockerfile adds a Dockerfile to the Parser
func (s *DockerfileParser) AddDockerfile(path string) {

	// Create a new Dockerfile entry
	dockerfile := Dockerfile{Path: path}
	cmds, err := df.ParseFile(path)

	// If we can't read for whatever reason, log the issue and continue
	if err != nil {
		log.Printf("%s is not a loadable Dockerfile, skipping.", path)
		return
	}

	// Add commands, parse FROMs, and LABELS
	dockerfile.AddCommands(cmds)
	dockerfile.UpdateFroms()
	s.Dockerfiles = append(s.Dockerfiles, dockerfile)
}

// Entrypoint to parse one or more Dockerfiles
func (s *DockerfileParser) Parse(path string) error {

	// Find Dockerfiles in path and allow prefixes
	paths, _ := utils.RecursiveFind(path, "Dockerfile", true)

	// Add each path as a Dockerfile to the parser to update
	for _, path = range paths {
		s.AddDockerfile(path)
	}

	// Keep track of updated count
	count := 0

	// Do we have updates? Count and write to file
	if s.CountUpdated() > 0 {
		for _, dockerfile := range s.Dockerfiles {
			if len(dockerfile.Updates) == 0 {
				continue
			}
			dockerfile.Write()
			count += 1
		}

	}
	fmt.Println("\n  ⭐️ Updated ⭐️")
	fmt.Printf("     Checked: %d\n", len(s.Dockerfiles))
	fmt.Printf("    Modified: %d\n", count)
	return nil
}
