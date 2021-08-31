package cli

import (
	"fmt"
	"github.com/DataDrake/cli-ng/v2/cmd"
	"github.com/vsoch/uptodate/version"
)

// Args and flags for generate
type VersionArgs struct {
}
type VersionFlags struct {
}

// Dockerfile updates one or more Dockerfile
var VersionCmd = cmd.Sub{
	Name:  "version",
	Alias: "v",
	Short: "Print the version to the terminal.",
	Flags: &VersionFlags{},
	Args:  &VersionArgs{},
	Run:   RunVersion,
}

func init() {
	cmd.Register(&VersionCmd)
}

// RunDockerfile updates one or more Dockerfile
func RunVersion(r *cmd.Root, c *cmd.Sub) {
	fmt.Println(version.Version)
}
