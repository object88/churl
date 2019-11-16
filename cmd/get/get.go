package get

import (
	"github.com/object88/churl/cmd/common"
	"github.com/object88/churl/cmd/get/latest"
	"github.com/object88/churl/cmd/traverse"
	"github.com/spf13/cobra"
)

// CreateCommand returns the intermediate 'get' subcommand
func CreateCommand(ca *common.CommonArgs) *cobra.Command {
	c := &cobra.Command{
		Use:   "get",
		Short: "get subcommands will return metadata about one or more charts",
	}

	c.AddCommand(
		latest.CreateCommand(ca),
	)

	return traverse.TraverseRunHooks(c)
}
