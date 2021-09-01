package cli

import (
	"github.com/DataDrake/cli-ng/v2/cmd"
	"github.com/vsoch/uptodate/version"
)

// GlobalFlags contains the flags for commands.
type GlobalFlags struct{}

// Root is the main command.
var Root *cmd.Root

// init create the root command and creates subcommands
func init() {
	Root = &cmd.Root{
		Name:      "uptodate",
		Short:     "Ensure repository assets are up to date.",
		Version:   version.Version,
		Copyright: "© 2021 Vanessa Sochat <@vsoch>",
		License:   "Licensed under the Apache License, Version 2.0",
	}
	cmd.Register(&cmd.Help)
	cmd.Register(&cmd.Version)
	cmd.Register(&cmd.GenManPages)
}
