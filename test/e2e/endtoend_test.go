//+build test_e2e

package e2e

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
	ctesting "github.com/object88/churl/internal/testing"
)

func Test_EndToEnd(t *testing.T) {
	wcd, teardown, err := NewWithChartMuseum(t)
	if err != nil {
		t.Fatalf("Failed to create test chart museum:\n%s", err.Error())
	}
	defer teardown()

	wcd.InstallChartMuseum()

	t.Log("Run start")
	wcd.Run()
	t.Log("Run complete")
}

func NewWithChartMuseum(t *testing.T) (*WithChartMuseum, func(), error) {
	testBinary := os.Getenv("TEST_BINARY_NAME")
	if testBinary == "" {
		t.Fatal("Environment variable '$TEST_BINARY_NAME' not provided, skipping.")
	}
	churl := findBinary(t, testBinary)
	helm := findBinary(t, "helm")
	kubectl := findBinary(t, "kubectl")

	wcd := &WithChartMuseum{
		t:             t,
		churlBinary:   churl,
		helmBinary:    helm,
		kubectlBinary: kubectl,
	}

	cancelFn := func() {
		err := wcd.Close()
		if err != nil {
			t.Errorf("Failed to close:\n%s", err.Error())
		}
	}

	return wcd, cancelFn, nil
}

type WithChartMuseum struct {
	t *testing.T

	churlBinary   *ctesting.Runnable
	helmBinary    *ctesting.Runnable
	kubectlBinary *ctesting.Runnable

	chartMuseumRepo string
	chartMuseum     string
	podName         string
	serviceName     string
}

func (wcd *WithChartMuseum) Run() {
	val := reflect.ValueOf(wcd)
	tval := val.Type()

	for i := 0; i < tval.NumMethod(); i++ {
		mval := tval.Method(i)
		if !strings.HasPrefix(mval.Name, "Test_") {
			continue
		}
		fn := mval.Func
		tfn := fn.Type()
		if tfn.NumOut() != 0 || tfn.NumIn() != 2 || tfn.In(1) != reflect.TypeOf(wcd.t) {
			continue
		}

		wcd.t.Run(mval.Name, func(t *testing.T) {
			t.Logf("Starting '%s'", mval.Name)

			// Run test preparation

			// Invoke the test
			fn.Call([]reflect.Value{val, reflect.ValueOf(t)})
			if !t.Failed() {
				// Exit early; test was successful
				return
			}

			// Run post-test-mortum
			t.Log("Test failed.")
		})
	}
}

// Close satisfies io.Close, used to clean up test chart museum
func (wcd *WithChartMuseum) Close() error {
	if wcd == nil {
		return nil
	}

	wcd.CleanupChartMuseum()

	return nil
}

func (wcd *WithChartMuseum) InstallChartMuseum() {
	chartMuseumRepo := fmt.Sprintf("churl-test-%s", uuid.New().String())
	chartMuseum := fmt.Sprintf("churl-test-cm-%s", uuid.New())

	_, exitcode := wcd.helmBinary.Run("repo", "add", chartMuseumRepo, "https://kubernetes-charts.storage.googleapis.com")
	if exitcode != 0 {
		wcd.t.Fatalf("Failed to add chart museum repo")
	}
	wcd.chartMuseumRepo = chartMuseumRepo

	_, exitcode = wcd.helmBinary.Run("install", chartMuseum, fmt.Sprintf("%s/chartmuseum", wcd.chartMuseumRepo), "--set", "env.open.DISABLE_API=false", "--wait")
	if exitcode != 0 {
		wcd.t.Fatalf("Failed to install chart museum")
	}

	wcd.chartMuseum = chartMuseum
	wcd.serviceName, wcd.podName = wcd.inspectMuseumDeployment()

	tmp, _ := ioutil.TempDir("", uuid.New().String())

	f := func(chart, version string) {
		sourcePackage := fmt.Sprintf("../charts/%s", chart)
		_, exitcode = wcd.helmBinary.Run("package", sourcePackage, "--destination", tmp, "--version", version)
		if exitcode != 0 {
			wcd.t.Fatalf("Failed to create chart tarball for chart '%s' version '%s'", chart, version)
		}
		chartFile := fmt.Sprintf("%s-%s.tgz", chart, version)
		sourceTarball := fmt.Sprintf("%s/%s", tmp, chartFile)
		destination := fmt.Sprintf("%s:/storage/%s", wcd.podName, chartFile)
		_, exitcode := wcd.kubectlBinary.Run("cp", sourceTarball, destination)
		if exitcode != 0 {
			wcd.t.Fatalf("Failed to copy into pod")
		}
	}
	f("foo", "1.0.1")
	f("foo", "1.1.0")
	f("bar", "2.0.0")
	f("bar", "2.0.2")
}

func (wcd *WithChartMuseum) CleanupChartMuseum() {
	if wcd.chartMuseum != "" {
		wcd.helmBinary.Run("delete", wcd.chartMuseum)
	}

	if wcd.chartMuseumRepo != "" {
		wcd.helmBinary.Run("repo", "remove", wcd.chartMuseumRepo)
	}
}

func (wcd *WithChartMuseum) generateDefaultManifest(chartpath string) string {
	const chartTemplate = `{
		"museums": [{
			"name": "default",
			"kubeContext": "%s",
			"serviceName": "%s",
			"namespace": "default",
			"port": "8080"
		}],
		"current": "default"
	}`

	err := os.MkdirAll(chartpath, 0644)
	if err != nil {
		wcd.t.Fatalf("Failed to create directory '%s' for chart", chartpath)
	}

	kubeContext, exitcode := wcd.kubectlBinary.Run("config", "current-context")
	if exitcode != 0 {
		wcd.t.Fatalf("Failed to get current kubectl context")
	}

	contents := fmt.Sprintf(chartTemplate, kubeContext, wcd.serviceName)

	chartfile := path.Join(chartpath, "manifest.json")
	err = ioutil.WriteFile(chartfile, []byte(contents), 0644)
	if err != nil {
		wcd.t.Fatalf("Failed to set up default chart")
	}

	return chartfile
}

func (wcd *WithChartMuseum) inspectMuseumDeployment() (string, string) {
	selector := fmt.Sprintf("release=%s", wcd.chartMuseum)

	serviceName, exitcode := wcd.kubectlBinary.Run("get", "services", "-l", selector, "--output", "name")
	if exitcode != 0 {
		wcd.t.Fatalf("Failed to find service name")
	}

	podName, exitcode := wcd.kubectlBinary.Run("get", "pods", "-l", selector, "--output", "name")
	if exitcode != 0 {
		wcd.t.Fatalf("Failed to find pod name")
	}

	// Need to trim off the "pod/" prefix.
	if strings.HasPrefix(podName, "pod/") {
		podName = podName[4:]
	}

	return strings.TrimSpace(serviceName), strings.TrimSpace(podName)
}
