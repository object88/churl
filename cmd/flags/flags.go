package flags

import (
	"os"
	"path"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	ConfigKey  string = "config"
	OutputKey         = "output"
	VerboseKey        = "verbose"
)

func CreateConfigFlag(flgs *pflag.FlagSet) {
	d, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}
	configFile := path.Join(d, "churl", "config.json")

	flgs.String(ConfigKey, configFile, "Path to configuration file")
	viper.BindPFlag(ConfigKey, flgs.Lookup(ConfigKey))
	viper.BindEnv(ConfigKey)
}
