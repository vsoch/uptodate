package cli

import (
	"fmt"
	"github.com/DataDrake/cli-ng/v2/cmd"
	"github.com/vsoch/uptodate/parsers"
	"github.com/vsoch/uptodate/utils"
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

	// If no root provided, assume parsing the PWD
	if len(args.Root) == 0 {
		args.Root = []string{utils.GetPwd()}
	}

	// Print the logo!
	fmt.Println(utils.GetLogo() + "                     dockerfile\n")

	// Update the dockerfiles with a Dockerfile parser
	parser := parsers.DockerfileParser{}
	parser.Parse(args.Root[0])
}
