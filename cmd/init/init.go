package init

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
}

// CreateCommand returns a new instance of a *cobra.Command for the init
// command
func CreateCommand(ca *common.CommonArgs) *cobra.Command {
	var c command
	c = command{
		Command: cobra.Command{
			Use: "init",
			RunE: func(cmd *cobra.Command, args []string) error {
				return c.Execute(cmd, args)
			},
		},
		CommonArgs: ca,
	}

	flgs := c.Flags()

	flags.CreateConfigFlag(flgs)

	return traverse.TraverseRunHooks(&c.Command)
}

func (c *command) Execute(cmd *cobra.Command, args []string) error {
	configFile := viper.GetString(flags.ConfigKey)
	configDir := path.Base(configFile)
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		return errors.Wrapf(err, "cannot create config directory '%s'", configDir)
	}

	m, err := manifest.Init(configFile)
	if err != nil {
		return errors.Wrapf(err, "failed to initialize churl manifest file at '%s'", configFile)
	}
	err = m.Save()
	if err != nil {
		return errors.Wrapf(err, "failed to save manifest file at '%s'", configFile)
	}

	defer m.Close()

	c.Logger.Infof("Created config file at '%s'", configFile)

	return nil
}
