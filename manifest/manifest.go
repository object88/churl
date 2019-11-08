package manifest

import "io"

// ChartMuseum describes the destination chart museum
type ChartMuseum struct {
	KubeContext string
	ServiceName string
	Port        uint64
}

// Manifest describes the configuration for the churl tool
type Manifest struct {
	Museums ChartMuseum
}

func Open(r io.Reader) (*Manifest, error) {
	return nil, nil
}

func OpenFromFile(manifestFilepath string) (*Manifest, error) {

	return Open(nil)
}
