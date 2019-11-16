package flags

import (
	"os"
	"path"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	// ConfigKey is used to specify where a churl config file can be found
	ConfigKey string = "config"

	// OutputKey determines the output format
	OutputKey = "output"

	// VerboseKey turns on verbose output to STDERR
	VerboseKey = "verbose"
)

// CreateConfigFlag adds the `--config` flag to the flagset, with an
// OS-specific default location
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

// CreateOutputFlag adds the `--output` flag to the flagset
func CreateOutputFlag(flgs *pflag.FlagSet) {
	var def Output
	flgs.String(OutputKey, def.String(), Values())
	viper.BindPFlag(OutputKey, flgs.Lookup(OutputKey))
	viper.BindEnv(OutputKey)
}

// ReadOutputFlag gets the specified output setting, and verifies that it is a
// legitimate value
func ReadOutputFlag() (Output, error) {
	raw := viper.GetString(OutputKey)
	var o Output
	if err := o.UnmarshalText([]byte(raw)); err != nil {
		return Unknown, err
	}
	return o, nil
}
