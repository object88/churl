package manifest

import (
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
)

// ChartMuseum describes the destination chart museum
type ChartMuseum struct {
	KubeContext string
	ServiceName string
	Namespace   string
	Port        string
}

type intermediateMuseum struct {
	*ChartMuseum
	name string
}

const (
	nameKey string = "name"

	kubeContextKey string = "kubeContext"
	namespaceKey          = "namespace"
	portKey               = "port"
	serviceNameKey        = "serviceName"
)

func (cm *ChartMuseum) validate() error {
	return nil
}

func (im *intermediateMuseum) UnmarshalJSON(b []byte) error {
	if im == nil {
		im = &intermediateMuseum{}
	}
	if im.ChartMuseum == nil {
		im.ChartMuseum = &ChartMuseum{}
	}

	var objMap map[string]*json.RawMessage
	err := json.Unmarshal(b, &objMap)
	if err != nil {
		return errors.Wrapf(err, "Failed to unmarshal top-level *intermediateMuseum")
	}

	extraKeys := map[string]*json.RawMessage{}
	foundKeys := map[string]struct{}{}
	for k, v := range objMap {
		if _, ok := foundKeys[k]; ok {
			return errors.Errorf("Found duplicate key '%s'", k)
		}
		foundKeys[k] = struct{}{}

		switch k {
		case kubeContextKey:
			err = json.Unmarshal(*v, &im.KubeContext)
		case nameKey:
			err = json.Unmarshal(*v, &im.name)
		case namespaceKey:
			err = json.Unmarshal(*v, &im.Namespace)
		case portKey:
			err = json.Unmarshal(*v, &im.Port)
		case serviceNameKey:
			err = json.Unmarshal(*v, &im.ServiceName)
		default:
			if _, ok := extraKeys[k]; ok {
				return errors.Errorf("Found duplicate (extraneous) key '%s'", k)
			}
			extraKeys[k] = v
		}
		if err != nil {
			return errors.Wrapf(err, "Internal error: failed to unmarshal property '%s'", k)
		}
	}

	if len(extraKeys) != 0 {
		var sb strings.Builder
		sb.WriteString("Found extra keys '")
		first := true
		for k := range extraKeys {
			if !first {
				sb.WriteString("', '")
			}
			sb.WriteString(k)
			first = false
		}
		sb.WriteString("'")
		return errors.Errorf(sb.String())
	}

	return nil
}
