package churl

import "strings"

import "encoding/json"

// ChurlVersion contains the version
var ChurlVersion = "unset"

// GitCommit contains the git SHA
var GitCommit = "unset"

type Version struct{}

// Churl returns the binary version
func (Version) Churl() string {
	return ChurlVersion
}

// Git returns the commit SHA of the source
func (Version) Git() string {
	return GitCommit
}

func (v Version) String() string {
	var sb strings.Builder
	sb.WriteString("Version:    ")
	sb.WriteString(v.Churl())
	sb.WriteRune('\n')
	sb.WriteString("Git commit: ")
	sb.WriteString(v.Git())
	sb.WriteRune('\n')
	return sb.String()
}

func (v Version) MarshalJSON() ([]byte, error) {
	x := struct {
		Version string `json:"version"`
		SHA     string `json:"sha"`
	}{
		Version: ChurlVersion,
		SHA:     GitCommit,
	}
	return json.Marshal(x)
}
