package cli

import (
	"fmt"
	"github.com/DataDrake/cli-ng/v2/cmd"
	"github.com/vsoch/uptodate/config"
	"github.com/vsoch/uptodate/parsers"
)

// Args and flags for generate
type DockerfileArgs struct {
	Root []string `zero:"true" desc:"A Dockerfile or directory to parse."`
}
type DockerfileFlags struct{}

// Dockerfile updates one or more Dockerfile
var Dockerfile = cmd.Sub{
	Name:  "dockerfile",
	Alias: "df",
	Short: "Update one or more Dockerfile.",
	Flags: &DockerfileFlags{},
	Args:  &DockerfileArgs{},
	Run:   RunDockerfile,
}

func init() {
	cmd.Register(&Dockerfile)
}

// RunDockerfile updates one or more Dockerfile
func RunDockerfile(r *cmd.Root, c *cmd.Sub) {

	args := c.Args.(*DockerfileArgs)
	flags := c.Flags.(*DockerfileFlags)

	// Create a new config to get envars (not used yet)
	c := config.NewConfig()

	// Update the dockerfiles with a Dockerfile parser
	parser := parsers.DockerfileParser{}
	parser.Parse(args.Root[0])
}
