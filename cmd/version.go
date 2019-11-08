package cmd

import (
	"fmt"
	"os"

	"github.com/object88/churl"
	"github.com/object88/churl/cmd/traverse"
	"github.com/spf13/cobra"
)

func createVersionCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "version",
		Short: "report the version of the tool",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(os.Stdout, "Churl version %s\n", churl.ChurlVersion)
			return nil
		},
	}

	return traverse.TraverseRunHooks(c)
}
