package init

import (
	"github.com/object88/churl/cmd/common"
	"github.com/object88/churl/cmd/traverse"
	"github.com/spf13/cobra"
)

type command struct {
	cobra.Command
	*common.CommonArgs
}

func CreateCommand(ca *common.CommonArgs) *cobra.Command {
	c := &command{
		Command: cobra.Command{
			Use: "init",
		},
		CommonArgs: ca,
	}

	return traverse.TraverseRunHooks(&c.Command)
}
