package cli

import (
	"fmt"
	"github.com/DataDrake/cli-ng/v2/cmd"
	"github.com/vsoch/uptodate/parsers/docker"
	"github.com/vsoch/uptodate/utils"
)

// Args and flags for generate
type DockerHierarchyArgs struct {
	Root []string `zero:"true" desc:"A docker hierarchy directory to parse."`
}
type DockerHierarchyFlags struct {
	DryRun bool `long:"dry-run" desc:"Preview changes but don't write."`
}

// DockerHierarchy updates a docker hierarchy
var DockerHierarchy = cmd.Sub{
	Name:  "dockerhierarchy",
	Alias: "dh",
	Short: "Update a docker hierarchy.",
	Flags: &DockerHierarchyFlags{},
	Args:  &DockerHierarchyArgs{},
	Run:   RunDockerHierarchy,
}

func init() {
	cmd.Register(&DockerHierarchy)
}

// RunDockerHierarchy updates a docker hierarchy
func RunDockerHierarchy(r *cmd.Root, c *cmd.Sub) {

	args := c.Args.(*DockerHierarchyArgs)
	flags := c.Flags.(*DockerHierarchyFlags)

	// If no root provided, assume parsing the PWD
	if len(args.Root) == 0 {
		args.Root = []string{utils.GetPwd()}
	}

	// Print the logo!
	fmt.Println(utils.GetLogo() + "               dockerhierarchy\n")

	// Update the docker hierarchy
	parser := docker.DockerHierarchyParser{}
	parser.Parse(args.Root[0], flags.DryRun)
}
