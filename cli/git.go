package cli

import (
	"fmt"
	"github.com/DataDrake/cli-ng/v2/cmd"
	"github.com/vsoch/uptodate/parsers/git"
	"github.com/vsoch/uptodate/utils"
)

// Args and flags for generate
type GitArgs struct {
	Root []string `zero:"true" desc:"A root directory to parse."`
}

type GitFlags struct{}

var Git = cmd.Sub{
	Name:  "git",
	Alias: "git",
	Short: "Get changed for current commit.",
	Flags: &GitFlags{},
	Args:  &GitArgs{},
	Run:   RunGit,
}

func init() {
	cmd.Register(&Git)
}

// RunGit to get changed files
func RunGit(r *cmd.Root, c *cmd.Sub) {

	args := c.Args.(*GitArgs)

	// If no root provided, assume parsing the PWD
	if len(args.Root) == 0 {
		args.Root = []string{utils.GetPwd()}
	}

	// Print the logo!
	fmt.Println(utils.GetLogo() + "                          git\n")

	// Update the dockerfiles with a Dockerfile parser
	parser := git.GitParser{}
	parser.Parse(args.Root[0])

}
