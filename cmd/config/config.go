package config

import (
	"github.com/object88/churl/cmd/common"
	"github.com/object88/churl/cmd/config/current"
	"github.com/object88/churl/cmd/traverse"
	"github.com/spf13/cobra"
)

func CreateCommand(ca *common.CommonArgs) *cobra.Command {
	c := &cobra.Command{
		Use:   "config",
		Short: "config subcommands will handle churl configuration",
	}

	c.AddCommand(
		current.CreateCommand(ca),
	)

	return traverse.TraverseRunHooks(c)
}
