package manifest

import (
	"encoding/json"
	"testing"
)

const knownGoodMuseum = `{
	"name": "default",
	"kubeContext": "krobot",
	"serviceName": "cm-chartmuseum",
	"port": 8080
}
`

func Test_Manifest_ChartMuseum_Unmarshal(t *testing.T) {
	im := &intermediateMuseum{}
	// var im intermediateMuseum

	err := json.Unmarshal([]byte(knownGoodMuseum), &im)
	if err != nil {
		t.Fatalf("Failed to unmarshal museum:\n%s", err.Error())
	}

	if im.KubeContext != "krobot" {
		t.Errorf("Failed to unmarshal KubeContext; expected '%s', actual '%s'", "krobot", im.KubeContext)
	}
}
