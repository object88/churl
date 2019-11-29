//+build test_e2e

package e2e

import (
	"io/ioutil"
	"path"
	"testing"

	"github.com/google/uuid"
)

func (wcd *WithChartMuseum) Test_Setup(t *testing.T) {
	chartdir, _ := ioutil.TempDir("", uuid.New().String())
	chartfile := path.Join(chartdir, "manifest.json")

	_, exitcode := wcd.churlBinary.Run("init", "--config", chartfile)
	if exitcode != 0 {
		t.Errorf("Init failed, exit code %d", exitcode)
	}

	// // TODO: not implemented; missing subcommand is invoked, but this still
	// // doesn't return a non-zero exit code.  Waiting on an update from
	// // spf13/cobra for a change to this behavior.
	// _, exitcode = wcd.churlBinary.Run("config", "add")
	// if exitcode != 0 {
	// 	t.Errorf("Init failed, exit code %d", exitcode)
	// }
}
