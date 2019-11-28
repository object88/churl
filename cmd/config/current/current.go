package current

import (
	"os"

	"github.com/object88/churl/cmd/common"
	"github.com/object88/churl/cmd/flags"
	"github.com/object88/churl/cmd/traverse"
	"github.com/object88/churl/manifest"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type command struct {
	cobra.Command
	*common.CommonArgs

	m *manifest.Manifest
}

// CreateCommand returns the 'current' subcommand
func CreateCommand(ca *common.CommonArgs) *cobra.Command {
	var c *command
	c = &command{
		Command: cobra.Command{
			Use:   "current",
			Short: "returns current configuration",
			PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
				return c.Preexecute(cmd, args)
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				return c.Execute(cmd, args)
			},
		},
		CommonArgs: ca,
	}

	flgs := c.Flags()

	// Config flag
	flags.CreateConfigFlag(flgs)

	return traverse.TraverseRunHooks(&c.Command)
}

func (c *command) Preexecute(cmd *cobra.Command, args []string) error {
	m, err := manifest.OpenFromFile(viper.GetString(flags.ConfigKey))
	if err != nil {
		return errors.Wrapf(err, "Failed to open manifest file")
	}
	c.m = m
	return nil
}

func (c *command) Execute(cmd *cobra.Command, args []string) error {
	_, err := c.m.WriteTo(os.Stdout)
	if err != nil {
		return errors.Wrapf(err, "Failed to write to STDOUT")
	}
	return nil
}
