package e2e

import (
	"fmt"
	"os"
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
	churl, ok, err := ctesting.NewRunnable(t, testBinary)
	if err != nil {
		t.Fatalf("Internal error finding helm binary:\n%s", err.Error())
	}
	if !ok {
		t.Fatalf("Failed to find 'churl' binary, cannot run tests")
	}

	helm, ok, err := ctesting.NewRunnable(t, "helm")
	if err != nil {
		t.Fatalf("Internal error finding helm binary:\n%s", err.Error())
	}
	if !ok {
		t.Fatalf("Failed to find 'helm' binary, cannot run tests")
	}

	wcd := &WithChartMuseum{
		t:           t,
		churlBinary: churl,
		helmBinary:  helm,
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

	churlBinary *ctesting.Runnable
	helmBinary  *ctesting.Runnable

	chartMuseumRepo string
	chartMuseum     string
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

	_, exitcode = wcd.helmBinary.Run("install", chartMuseum, fmt.Sprintf("%s/chartmuseum", wcd.chartMuseumRepo))
	if exitcode != 0 {
		wcd.t.Fatalf("Failed to install add chart museum")
	}
	wcd.chartMuseum = chartMuseum
}

func (wcd *WithChartMuseum) CleanupChartMuseum() {
	if wcd.chartMuseum != "" {
		wcd.helmBinary.Run("delete", wcd.chartMuseum)
	}

	if wcd.chartMuseumRepo != "" {
		wcd.helmBinary.Run("repo", "remove", wcd.chartMuseumRepo)
	}
}
