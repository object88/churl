package current

import (
	"os"
	"path"

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

	d, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}
	configFile := path.Join(d, "churl", "config.json")

	flgs := c.Flags()

	flgs.String(flags.ConfigKey, configFile, "Path to configuration file")
	viper.BindPFlag(flags.ConfigKey, flgs.Lookup(flags.ConfigKey))
	viper.BindEnv(flags.ConfigKey)

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
