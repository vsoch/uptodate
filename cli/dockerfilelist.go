package cli

import (
	"github.com/DataDrake/cli-ng/v2/cmd"
	"github.com/vsoch/uptodate/parsers/docker"
	"github.com/vsoch/uptodate/utils"
)

// Args and flags for generate
type DockerfileListArgs struct {
	Root []string `zero:"true" desc:"A Dockerfile or directory to parse."`
}

type DockerfileListFlags struct {
	NoEmptyArgs   bool   `long:"no-empty-build-args" desc:"Do not include Dockerfile with empty build args (defaults to false)"`
	NoIncludeArgs bool   `long:"no-build-args" desc:"Do not include Dockerfile with any build args (defaults to false)"`
	Changes       bool   `long:"changes" desc:"Only consider changed uptodate files"`
	Branch        string `long:"branch" desc:"Branch to compare HEAD against, defaults to main"`
}

// Dockerfile updates one or more Dockerfile
var DockerfileList = cmd.Sub{
	Name:  "dockerfilelist",
	Alias: "dl",
	Short: "List one or more Dockerfile.",
	Flags: &DockerfileListFlags{},
	Args:  &DockerfileListArgs{},
	Run:   RunDockerfileList,
}

func init() {
	cmd.Register(&DockerfileList)
}

// RunDockerfile updates one or more Dockerfile
func RunDockerfileList(r *cmd.Root, c *cmd.Sub) {

	args := c.Args.(*DockerfileListArgs)
	flags := c.Flags.(*DockerfileListFlags)

	// If no root provided, assume parsing the PWD
	if len(args.Root) == 0 {
		args.Root = []string{utils.GetPwd()}
	}

	// Set default branch
	if flags.Branch == "" {
		flags.Branch = "main"
	}

	// Update the dockerfiles with a Dockerfile parser
	parser := docker.DockerfileListParser{}
	parser.Parse(args.Root, !flags.NoEmptyArgs, !flags.NoIncludeArgs, flags.Changes, flags.Branch)
}
