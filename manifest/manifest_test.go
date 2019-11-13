package manifest

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/google/uuid"
	jmespath "github.com/jmespath/go-jmespath"
)

const knownGoodManifest = `{
	"museums": [
		{
			"name": "default",
			"kubeContext": "krobot",
			"serviceName": "cm-chartmuseum",
			"port": 8080
		}
	]
}
`

func Test_Manifest_Open(t *testing.T) {
	r := strings.NewReader(knownGoodManifest)

	m, err := Open(r)
	if err != nil {
		t.Fatalf("Failed to open manifest string:\n%s", err.Error())
	}

	if m == nil {
		t.Fatalf("Failed to get museum back from open")
	}
}

func Test_Manifest_OpenFromFile(t *testing.T) {
	tempf := writeManifestFile(t, knownGoodManifest)
	defer os.Remove(tempf)

	m, err := OpenFromFile(tempf)
	if err != nil {
		t.Fatalf("Failed to open manifest file:\n%s", err.Error())
	}

	if len(m.Museums) != 1 {
		t.Errorf("Incorrect number of museums: expected 1, actual %d", len(m.Museums))
	}

	cm, ok := m.Museums["default"]
	if !ok {
		t.Errorf("Failed to find chart museum 'default'")
	}

	if cm.KubeContext != "krobot" {
		t.Errorf("Chart museum did not unmarshal")
	}
}

func Test_Manifest_Save(t *testing.T) {
	chartdir, _ := ioutil.TempDir("", uuid.New().String())
	chartfile := path.Join(chartdir, "manifest.json")
	m, err := Init(chartfile)
	if err != nil {
		t.Fatalf("Failed to init '%s':\n%s", chartfile, err.Error())
	}

	m.Museums["foo"] = &ChartMuseum{
		Port: 123,
	}
	m.Museums["bar"] = &ChartMuseum{
		KubeContext: "blah",
	}

	err = m.Save()
	if err != nil {
		t.Fatalf("Failed to save:\n%s", err.Error())
	}

	m.Close()

	t.Logf("Chart file: '%s'", chartfile)

	f, err := os.Open(chartfile)
	if err != nil {
		t.Fatalf("Failed to read manifest file:\n%s", err.Error())
	}
	defer f.Close()

	query := "museums"
	jmes, err := jmespath.Compile(query)
	if err != nil {
		t.Fatalf("Failed to compile query '%s':\n%s", query, err.Error())
	}

	var data interface{}
	dec := json.NewDecoder(f)
	dec.Decode(&data)

	result, err := jmes.Search(data)
	if err != nil {
		t.Fatalf("Failed to execute jmes query with query '%s':\n%s", query, err.Error())
	}

	if result == nil {
		t.Errorf("Did not find JSON element with query '%s'", query)
	}

	t.Errorf("NOTOK")
}

func Test_Manifest_Close(t *testing.T) {
	tcs := []struct {
		name string
		m    *Manifest
	}{
		{
			name: "from file",
			m:    func() *Manifest { m, _ := OpenFromFile(writeManifestFile(t, knownGoodManifest)); return m }(),
		},
		{
			name: "from memory",
			m:    func() *Manifest { m, _ := Open(strings.NewReader(knownGoodManifest)); return m }(),
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.m.Close()
			if err != nil {
				t.Errorf("Unexpected error while closing:\n%s", err.Error())
			}
		})
	}
}

func writeManifestFile(t *testing.T, contents string) string {
	tempf, err := ioutil.TempFile("", "manifest.*.json")
	if err != nil {
		t.Fatalf("Failed to create temporary file:\n%s", err.Error())
	}
	defer tempf.Close()

	s := tempf.Name()

	_, err = tempf.Write([]byte(contents))
	if err != nil {
		t.Fatalf("Failed to write to temporary file:\n%s", err.Error())
	}

	return s
}
