package version

import (
	"os"

	"github.com/object88/churl"
	"github.com/object88/churl/cmd/traverse"
	"github.com/spf13/cobra"
)

type command struct {
	cobra.Command
}

// CreateCommand returns the version command
func CreateCommand() *cobra.Command {
	var c *command
	c = &command{
		Command: cobra.Command{
			Use:   "version",
			Short: "report the version of the tool",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				var v churl.Version
				os.Stdout.WriteString(v.String())
				return nil
			},
		},
	}

	return traverse.TraverseRunHooks(&c.Command)
}
