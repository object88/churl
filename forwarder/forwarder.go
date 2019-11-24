package forwarder

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/object88/churl/manifest"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/polymorphichelpers"
	"k8s.io/kubectl/pkg/scheme"
	"k8s.io/kubectl/pkg/util"
)

type Forwarder struct {
	fw *portforward.PortForwarder

	stop chan struct{}

	SourcePort string
}

func Open(factory cmdutil.Factory, config *rest.Config, m *manifest.Manifest, options ...Option) (*Forwarder, error) {
	o := NewOptions()
	for _, opt := range options {
		if err := opt(o); err != nil {
			return nil, errors.Wrapf(err, "Option invalid")
		}
	}

	cm := m.Current()

	builder := factory.NewBuilder().
		WithScheme(scheme.Scheme, scheme.Scheme.PrioritizedVersionsAllGroups()...).
		ContinueOnError().
		NamespaceParam(cm.Namespace).DefaultNamespace()
	builder.ResourceNames("pods", cm.ServiceName)

	obj, err := builder.Do().Object()
	if err != nil {
		return nil, err
	}

	forwardablePod, err := polymorphichelpers.AttachablePodForObjectFn(factory, obj, o.podTimeout)
	if err != nil {
		return nil, err
	}

	sourcePort := "9999"
	destinationPort := cm.Port

	// handle service port mapping to target port if needed
	switch t := obj.(type) {
	case *v1.Service:
		sourcePort, destinationPort, err = translateServicePortToTargetPort(sourcePort, destinationPort, *t, *forwardablePod)
	default:
		sourcePort, destinationPort, err = convertPodNamedPortToNumber(sourcePort, destinationPort, *forwardablePod)
	}
	if err != nil {
		return nil, err
	}

	transport, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		return nil, err
	}

	portmap := []string{fmt.Sprintf("%s:%s", sourcePort, destinationPort)}

	client, err := factory.RESTClient()
	if err != nil {
		return nil, err
	}

	req := client.Post().Resource("pods").Namespace(forwardablePod.Namespace).Name(forwardablePod.Name).SubResource("portforward")
	stop := make(chan struct{})

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", req.URL())
	fw, err := portforward.New(dialer, portmap, stop, o.ready, o.out, o.err)
	if err != nil {
		return nil, err
	}

	f := &Forwarder{
		fw:         fw,
		SourcePort: sourcePort,
		stop:       stop,
	}

	return f, nil
}

func (f *Forwarder) ForwardPorts() error {
	return f.fw.ForwardPorts()
}

func (f *Forwarder) Close() error {
	close(f.stop)
	f.fw.Close()

	return nil
}

// Translates service port to target port
// It rewrites ports as needed if the Service port declares targetPort.
// It returns an error when a named targetPort can't find a match in the pod, or the Service did not declare
// the port.
func translateServicePortToTargetPort(localPort, remotePort string, svc v1.Service, pod v1.Pod) (string, string, error) {
	portnum, err := strconv.Atoi(remotePort)
	if err != nil {
		svcPort, err := util.LookupServicePortNumberByName(svc, remotePort)
		if err != nil {
			return "", "", err
		}
		portnum = int(svcPort)

		if localPort == remotePort {
			localPort = strconv.Itoa(portnum)
		}
	}
	containerPort, err := util.LookupContainerPortNumberByServicePort(svc, pod, int32(portnum))
	if err != nil {
		// can't resolve a named port, or Service did not declare this port, return an error
		return "", "", err
	}

	if int32(portnum) != containerPort {
		return localPort, strconv.Itoa(int(containerPort)), nil
	}

	return localPort, remotePort, nil
}

// convertPodNamedPortToNumber converts named ports into port numbers
// It returns an error when a named port can't be found in the pod containers
func convertPodNamedPortToNumber(localPort, remotePort string, pod v1.Pod) (string, string, error) {
	containerPortStr := remotePort
	_, err := strconv.Atoi(remotePort)
	if err != nil {
		containerPort, err := util.LookupContainerPortNumberByName(pod, remotePort)
		if err != nil {
			return "", "", err
		}

		containerPortStr = strconv.Itoa(int(containerPort))
	}

	return localPort, containerPortStr, nil
}
