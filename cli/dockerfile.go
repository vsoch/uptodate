package cli

import (
	"fmt"
	"github.com/DataDrake/cli-ng/v2/cmd"
	"github.com/vsoch/uptodate/parsers/docker"
	"github.com/vsoch/uptodate/utils"
)

// Args and flags for generate
type DockerfileArgs struct {
	Root []string `zero:"true" desc:"A Dockerfile or directory to parse."`
}
type DockerfileFlags struct {
	DryRun  bool   `long:"dry-run" desc:"Preview changes but don't write."`
	Changes bool   `long:"changes" desc:"Only consider changed uptodate files"`
	Branch  string `long:"branch" desc:"Branch to compare HEAD against, defaults to main"`
}

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

	// If no root provided, assume parsing the PWD
	if len(args.Root) == 0 {
		args.Root = []string{utils.GetPwd()}
	}

	// Set default branch
	if flags.Branch == "" {
		flags.Branch = "main"
	}

	// Print the logo!
	fmt.Println(utils.GetLogo() + "                     dockerfile\n")

	// Update the dockerfiles with a Dockerfile parser
	parser := docker.DockerfileParser{}
	parser.Parse(args.Root[0], flags.DryRun, flags.Changes, flags.Branch)

}
