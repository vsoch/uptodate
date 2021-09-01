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

	// If no root provided, assume parsing the PWD
	if len(args.Root) == 0 {
		args.Root = []string{utils.GetPwd()}
	}

	// Update the dockerfiles with a Dockerfile parser
	parser := docker.DockerfileListParser{}
	parser.Parse(args.Root[0])

}
