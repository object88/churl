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

func CreateOutputFlag(flgs *pflag.FlagSet) {
	var def Output
	flgs.String(OutputKey, def.String(), Values())
	viper.BindPFlag(OutputKey, flgs.Lookup(OutputKey))
	viper.BindEnv(OutputKey)
}

func ReadOutputFlag() (Output, error) {
	raw := viper.GetString(OutputKey)
	var o Output
	if err := o.UnmarshalText([]byte(raw)); err != nil {
		return Unknown, err
	}
	return o, nil
}
