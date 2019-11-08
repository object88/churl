package latest

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/object88/churl/cmd/common"
	"github.com/object88/churl/cmd/traverse"
	"github.com/object88/churl/forwarder"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes/scheme"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/polymorphichelpers"
	"k8s.io/kubectl/pkg/util"
)

const (
	// Amount of time to wait until at least one pod is running
	defaultPodPortForwardWaitTimeout = 2 * time.Second
)

type LatestCommand struct {
	cobra.Command
	*common.CommonArgs

	namespace string

	cflags *genericclioptions.ConfigFlags

	f     *forwarder.Forwarder
	ready chan struct{}
}

func CreateCommand(ca *common.CommonArgs) *cobra.Command {
	var lc *LatestCommand

	lc = &LatestCommand{
		Command: cobra.Command{
			Use:   "latest",
			Short: "latest will return the metadata for the newest version of a chart",
			// Args:  cobra.NoArgs,
			PreRunE: func(cmd *cobra.Command, args []string) error {
				return lc.Preexecute(cmd, args)
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				return lc.Execute(cmd, args)
			},
		},
		CommonArgs: ca,
		ready:      make(chan struct{}),
	}

	flags := lc.Flags()

	lc.cflags = genericclioptions.NewConfigFlags(false)
	lc.cflags.AddFlags(flags)

	cmdutil.AddPodRunningTimeoutFlag(&lc.Command, defaultPodPortForwardWaitTimeout)

	return traverse.TraverseRunHooks(&lc.Command)
}

func (lc *LatestCommand) Preexecute(cmd *cobra.Command, args []string) error {
	config, err := lc.cflags.ToRESTConfig()
	if err != nil {
		return errors.Wrapf(err, "Failed to get REST config")
	}

	f := cmdutil.NewFactory(lc.cflags)

	lc.namespace, _, err = f.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return err
	}

	builder := f.NewBuilder().
		WithScheme(scheme.Scheme, scheme.Scheme.PrioritizedVersionsAllGroups()...).
		ContinueOnError().
		NamespaceParam(lc.namespace).DefaultNamespace()

	getPodTimeout, err := cmdutil.GetPodRunningTimeoutFlag(cmd)
	if err != nil {
		return cmdutil.UsageErrorf(cmd, err.Error())
	}

	resourceName := args[0]
	builder.ResourceNames("pods", resourceName)

	obj, err := builder.Do().Object()
	if err != nil {
		return err
	}

	forwardablePod, err := polymorphichelpers.AttachablePodForObjectFn(f, obj, getPodTimeout)
	if err != nil {
		return err
	}

	// o.PodName = forwardablePod.Name

	var ports []string

	incomingPorts := []string{"9999:8080"}

	// handle service port mapping to target port if needed
	switch t := obj.(type) {
	case *v1.Service:
		ports, err = translateServicePortToTargetPort(incomingPorts, *t, *forwardablePod)
		if err != nil {
			return err
		}
	default:
		ports, err = convertPodNamedPortToNumber(incomingPorts, *forwardablePod)
		if err != nil {
			return err
		}
	}

	fmt.Printf("ports: %s\n", ports)

	// clientset, err := f.KubernetesClientSet()
	// if err != nil {
	// 	return err
	// }

	// o.PodClient = clientset.CoreV1()

	lc.f, err = forwarder.Open(f, config, forwardablePod, forwarder.Out(os.Stdout), forwarder.Err(os.Stderr), forwarder.Ready(lc.ready))
	if err != nil {
		return errors.Wrapf(err, "Failed to open forwarder")
	}

	go func() {
		err = lc.f.ForwardPorts()
		if err != nil {
			fmt.Printf("Nope: %s\n", err.Error())
		}
	}()

	fmt.Printf("Waiting for port forward to be ready...\n")

	return nil
}

func (lc *LatestCommand) Execute(cmd *cobra.Command, args []string) error {
	<-lc.ready

	lc.Logger.Infof("Ready\n")

	resp, err := http.Get("http://localhost:9999/index.yaml")
	if err != nil {
		return errors.Wrapf(err, "Failed to make request")
	}
	defer resp.Body.Close()

	fmt.Printf("Status: %s\n", resp.Status)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrapf(err, "Failed to read body")
	}

	fmt.Printf("Body:\n%s\n", body)

	return nil
}

// splitPort splits port string which is in form of [LOCAL PORT]:REMOTE PORT
// and returns local and remote ports separately
func splitPort(port string) (local, remote string) {
	parts := strings.Split(port, ":")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}

	return parts[0], parts[0]
}

// Translates service port to target port
// It rewrites ports as needed if the Service port declares targetPort.
// It returns an error when a named targetPort can't find a match in the pod, or the Service did not declare
// the port.
func translateServicePortToTargetPort(ports []string, svc v1.Service, pod v1.Pod) ([]string, error) {
	var translated []string
	for _, port := range ports {
		localPort, remotePort := splitPort(port)

		portnum, err := strconv.Atoi(remotePort)
		if err != nil {
			svcPort, err := util.LookupServicePortNumberByName(svc, remotePort)
			if err != nil {
				return nil, err
			}
			portnum = int(svcPort)

			if localPort == remotePort {
				localPort = strconv.Itoa(portnum)
			}
		}
		containerPort, err := util.LookupContainerPortNumberByServicePort(svc, pod, int32(portnum))
		if err != nil {
			// can't resolve a named port, or Service did not declare this port, return an error
			return nil, err
		}

		if int32(portnum) != containerPort {
			translated = append(translated, fmt.Sprintf("%s:%d", localPort, containerPort))
		} else {
			translated = append(translated, port)
		}
	}
	return translated, nil
}

// convertPodNamedPortToNumber converts named ports into port numbers
// It returns an error when a named port can't be found in the pod containers
func convertPodNamedPortToNumber(ports []string, pod v1.Pod) ([]string, error) {
	var converted []string
	for _, port := range ports {
		localPort, remotePort := splitPort(port)

		containerPortStr := remotePort
		_, err := strconv.Atoi(remotePort)
		if err != nil {
			containerPort, err := util.LookupContainerPortNumberByName(pod, remotePort)
			if err != nil {
				return nil, err
			}

			containerPortStr = strconv.Itoa(int(containerPort))
		}

		if localPort != remotePort {
			converted = append(converted, fmt.Sprintf("%s:%s", localPort, containerPortStr))
		} else {
			converted = append(converted, containerPortStr)
		}
	}

	return converted, nil
}
