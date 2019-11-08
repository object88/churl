package forwarder

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"k8s.io/kubectl/pkg/cmd/util"
)

type Forwarder struct {
	fw *portforward.PortForwarder

	stop chan struct{}
}

func Open(factory util.Factory, config *rest.Config, pod *v1.Pod, options ...Option) (*Forwarder, error) {
	transport, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		return nil, err
	}

	o := NewOptions()
	for _, opt := range options {
		if err := opt(o); err != nil {
			return nil, errors.Wrapf(err, "Option invalid")
		}
	}

	client, err := factory.RESTClient()
	if err != nil {
		return nil, err
	}

	req := client.Post().Resource("pods").Namespace(pod.Namespace).Name(pod.Name).SubResource("portforward")
	stop := make(chan struct{})

	fmt.Printf("Port forward target: '%s'\n", req.URL())

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", req.URL())
	fw, err := portforward.New(dialer, []string{"9999:8080"}, stop, o.ready, o.out, o.err)
	if err != nil {
		return nil, err
	}

	f := &Forwarder{
		fw:   fw,
		stop: stop,
	}

	fmt.Printf("Have forwarder: %#v\n", f)

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
