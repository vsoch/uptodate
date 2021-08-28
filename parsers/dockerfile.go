package parsers

// The Dockerfile parser is optimized to find and capture FROM statements,
// and then keep track of LABELs

import (
	//	"bufio"
	"fmt"
	"log"
	"strings"

	"github.com/vsoch/uptodate/utils"
	"path/filepath"
	//	"github.com/DataDrake/cuppa/results"
	//	"github.com/DataDrake/cuppa/version"
	//	lookout "github.com/alecbcs/lookout/update"
	df "github.com/asottile/dockerfile"
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
	Path string
	Raw  string
	Cmds map[string][]Command // Lookup by command type for quicker parsing
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
	// TODO stopped here, we will want to run lookout here!
	for _, cmd := range cmds {
		d.AddCommand(cmd)
	}
}

// AddCommand parses a df.Command into a uptodate Command and adds to list
func (d *Dockerfile) AddCommand(cmd df.Command) {
	extendedCmd := Command(cmd)

	// We only care about FROM and LABEL (for now)
	commandType := strings.ToLower(cmd.Cmd)
	if utils.IncludesString(commandType, []string{"from", "label"}) {

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

	// Loop through FROMs and update!
	for _, from := range d.Cmds["from"] {
		fmt.Println(from)
	}

}

// DockerfileParser holds one or more Dockerfile
type DockerfileParser struct {
	Dockerfiles []Dockerfile
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
	fmt.Println(dockerfile.Cmds)
	dockerfile.UpdateFroms()
}

// Entrypoint to parse one or more Dockerfiles
func (s *DockerfileParser) Parse(path string) error {

	// Find Dockerfiles in path and allow prefixes
	paths, _ := utils.RecursiveFind(path, "Dockerfile", true)

	// Parse each path into a Dockerfile
	for _, path = range paths {

		// Add the dockerfile to the parser
		s.AddDockerfile(path)

	}
	return nil
}
