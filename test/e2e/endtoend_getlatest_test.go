package e2e

import (
	"io/ioutil"
	"testing"

	"github.com/google/uuid"
)

func (wcd *WithChartMuseum) Test_GetLatest(t *testing.T) {
	chartdir, _ := ioutil.TempDir("", uuid.New().String())
	chartfile := wcd.generateDefaultManifest(chartdir)

	result, exitcode := wcd.churlBinary.Run("get", "latest", "foo", "--config", chartfile)
	if exitcode != 0 {
		t.Errorf("Failed to get latest: exit code %d", exitcode)
	}

	t.Logf(result)
	// t.Errorf("")
}
