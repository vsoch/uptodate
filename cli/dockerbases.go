package cli

import (
	"fmt"
	"os"

	"github.com/DataDrake/cli-ng/v2/cmd"
	"github.com/vsoch/uptodate/parsers/docker"
	"github.com/vsoch/uptodate/utils"
)

// Args and flags for generate
type DockerBasesArgs struct {
	Root []string `zero:"true" desc:"A root directory to parse."`
}

type DockerBasesFlags struct {
	Bases    string `long:"bases" desc:"The directory of bases to build (required)."`
	Registry string `long:"registry" desc:"A registry is required to look up changes from previous builds."`
	Changes  bool   `long:"changes" desc:"Only consider changed uptodate files"`
	All      bool   `long:"all" desc:"Build all entries in the matrix regardless of changes"`
	Branch   string `long:"branch" desc:"Branch to compare HEAD against, defaults to main"`
}

// Dockerfile updates one or more Dockerfile
var DockerBases = cmd.Sub{
	Name:  "dockerbases",
	Alias: "db",
	Short: "Generate a matrix of builds with common docker bases",
	Flags: &DockerBasesFlags{},
	Args:  &DockerBasesArgs{},
	Run:   RunDockerBases,
}

func init() {
	cmd.Register(&DockerBases)
}

// RunDockerfile updates one or more Dockerfile
func RunDockerBases(r *cmd.Root, c *cmd.Sub) {

	args := c.Args.(*DockerBasesArgs)
	flags := c.Flags.(*DockerBasesFlags)

	// If no root provided, assume parsing the PWD
	if len(args.Root) == 0 {
		args.Root = []string{utils.GetPwd()}
	}

	if flags.Bases == "" {
		fmt.Println("Please provide a --bases root.")
		os.Exit(1)
	}
	// Print the logo!
	fmt.Println(utils.GetLogo() + "                     dockerbases\n")

	// Set default branch
	if flags.Branch == "" {
		flags.Branch = "main"
	}

	// Update the dockerfiles with a Dockerfile parser
	parser := docker.DockerBasesParser{}
	parser.Parse(flags.Bases, args.Root, flags.Changes, flags.Branch, flags.Registry, flags.All)

}
