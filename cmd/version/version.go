package version

import (
	"encoding/json"
	"os"

	"github.com/object88/churl"
	"github.com/object88/churl/cmd/flags"
	"github.com/object88/churl/cmd/traverse"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

				switch viper.GetString(flags.OutputKey) {
				case "text":
					os.Stdout.WriteString(v.String())
				case "json":
					enc := json.NewEncoder(os.Stdout)
					enc.SetIndent("", "  ")
					err := enc.Encode(v)
					if err != nil {
						return errors.Wrapf(err, "internal error: failed to encode version")
					}
				case "json-compact":
					enc := json.NewEncoder(os.Stdout)
					err := enc.Encode(v)
					if err != nil {
						return errors.Wrapf(err, "internal error: failed to encode version")
					}
				}

				return nil
			},
		},
	}

	flgs := c.Flags()

	flgs.String(flags.OutputKey, "text", "Output format ")
	viper.BindPFlag(flags.OutputKey, flgs.Lookup(flags.OutputKey))
	viper.BindEnv(flags.OutputKey)

	return traverse.TraverseRunHooks(&c.Command)
}
