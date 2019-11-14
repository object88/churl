package version

import (
	"testing"

	ctesting "github.com/object88/churl/cmd/internal/testing"
)

func Test_Cmd_Version(t *testing.T) {
	out := ctesting.Run(t, "version")
	if out == "Churl version unset\n" {
		t.Errorf("Version is not set: '%s'", out)
	}
}
