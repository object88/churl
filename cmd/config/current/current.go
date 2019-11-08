package current

import (
	"github.com/object88/churl/cmd/common"
	"github.com/object88/churl/cmd/traverse"
	"github.com/spf13/cobra"
)

func CreateCommand(ca *common.CommonArgs) *cobra.Command {
	c := &cobra.Command{
		Use:   "current",
		Short: "returns current configuration",
	}

	return traverse.TraverseRunHooks(c)
}
