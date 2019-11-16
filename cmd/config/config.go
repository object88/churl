package config

import (
	"github.com/object88/churl/cmd/common"
	"github.com/object88/churl/cmd/config/current"
	"github.com/object88/churl/cmd/traverse"
	"github.com/spf13/cobra"
)

type command struct {
	cobra.Command
	*common.CommonArgs
}

// CreateCommand returns the intermediate 'config' subcommand
func CreateCommand(ca *common.CommonArgs) *cobra.Command {
	var c *command
	c = &command{
		Command: cobra.Command{
			Use:   "config",
			Short: "config subcommands will handle churl configuration",
		},
		CommonArgs: ca,
	}

	c.AddCommand(
		current.CreateCommand(ca),
	)

	return traverse.TraverseRunHooks(&c.Command)
}
