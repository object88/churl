package latest

import (
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/object88/churl"
	"github.com/object88/churl/cmd/common"
	"github.com/object88/churl/cmd/flags"
	"github.com/object88/churl/cmd/traverse"
	"github.com/object88/churl/forwarder"
	"github.com/object88/churl/manifest"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

const (
	// Amount of time to wait until at least one pod is running
	defaultPodPortForwardWaitTimeout = 2 * time.Second
)

type command struct {
	cobra.Command
	*common.CommonArgs

	m *manifest.Manifest

	cflags *genericclioptions.ConfigFlags

	f     *forwarder.Forwarder
	ready chan struct{}

	chartpath string
	meta      *churl.MetadataReader
}

func CreateCommand(ca *common.CommonArgs) *cobra.Command {
	var c *command

	c = &command{
		Command: cobra.Command{
			Use:   "latest CHARTNAME",
			Short: "latest will return the metadata for the newest version of a chart",
			Args:  cobra.MinimumNArgs(1),
			PreRunE: func(cmd *cobra.Command, args []string) error {
				return c.Preexecute(cmd, args)
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				return c.Execute(cmd, args)
			},
			PostRunE: func(cmd *cobra.Command, args []string) error {
				return c.Postexecute(cmd, args)
			},
		},
		CommonArgs: ca,
		ready:      make(chan struct{}),
	}

	flgs := c.Flags()

	// Config flag
	flags.CreateConfigFlag(flgs)

	c.cflags = genericclioptions.NewConfigFlags(false)
	c.cflags.Namespace = nil
	c.cflags.AddFlags(flgs)

	cmdutil.AddPodRunningTimeoutFlag(&c.Command, defaultPodPortForwardWaitTimeout)

	return traverse.TraverseRunHooks(&c.Command)
}

func (c *command) Preexecute(cmd *cobra.Command, args []string) error {
	for k, v := range args {
		args[k] = strings.TrimSpace(v)
	}
	c.chartpath = strings.Join(args, "/")

	// Open the manifest file
	m, err := manifest.OpenFromFile(viper.GetString(flags.ConfigKey))
	if err != nil {
		return errors.Wrapf(err, "Failed to open manifest file")
	}
	c.m = m

	config, err := c.cflags.ToRESTConfig()
	if err != nil {
		return errors.Wrapf(err, "Failed to get REST config")
	}

	f := cmdutil.NewFactory(c.cflags)

	// Get timeout from cobra.Command
	podTimeout, err := cmdutil.GetPodRunningTimeoutFlag(cmd)
	if err != nil {
		return cmdutil.UsageErrorf(cmd, err.Error())
	}

	// Set up options and open port forward
	options := []forwarder.Option{
		forwarder.Out(os.Stderr),
		forwarder.Err(os.Stderr),
		forwarder.PodTimeout(podTimeout),
		forwarder.Ready(c.ready),
	}
	c.f, err = forwarder.Open(f, config, c.m, options...)
	if err != nil {
		return errors.Wrapf(err, "Failed to open forwarder")
	}

	go func() {
		err = c.f.ForwardPorts()
		if err != nil {
			c.Logger.Errorf("Failed to forward port: %s\n", err.Error())
		}
	}()

	c.Logger.Infof("Waiting for port forward to be ready...\n")

	// Set up requester into cm pod
	c.meta, err = churl.NewMetadataReader(c.f.SourcePort)
	if err != nil {
		return errors.Wrapf(err, "Internal error: failed to create metadata reader")
	}

	return nil
}

func (c *command) Execute(cmd *cobra.Command, args []string) error {
	// Wait until the forwarder is ready
	<-c.ready

	c.Logger.Infof("Ready\n")

	result, err := c.meta.Do(c.chartpath)
	if err != nil {
		return errors.Wrapf(err, "Could not get get chart at '%s'", c.chartpath)
	}

	enc := json.NewEncoder(os.Stdout)
	err = enc.Encode(result)
	if err != nil {
		return errors.Wrapf(err, "Internal error: failed to encode returned chart")
	}

	return nil
}

func (c *command) Postexecute(cmd *cobra.Command, args []string) error {
	if c == nil {
		return nil
	}

	if c.m != nil {
		c.m.Close()
		c.m = nil
	}

	return nil
}
