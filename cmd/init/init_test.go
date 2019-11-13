package init

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmespath/go-jmespath"
	"github.com/pkg/errors"
)

func Test_Init(t *testing.T) {
	chartdir, _ := ioutil.TempDir("", uuid.New().String())
	chartfile := path.Join(chartdir, "manifest.json")

	run(t, "init", "--config", chartfile)

	validateManifest(t, chartfile)
}

func run(t *testing.T, args ...string) {
	bin := os.Getenv("TEST_BINARY_NAME")
	if bin == "" {
		t.Skipf("Environment variable '$TEST_BINARY_NAME' not provided, skipping.")
	}
	binaryPath, err := exec.LookPath(bin)
	if err != nil {
		t.Fatalf("Environment variable '$TEST_BINARY_NAME' has value '%s', but could not be found:\n%s", bin, err.Error())
	}

	ctx := context.Background()
	if true {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
	}
	cmd := exec.CommandContext(context.Background(), binaryPath, args...)

	pout, pin := io.Pipe()
	cmd.Stdout = pin
	cmd.Stderr = pin

	go func() {
		scanner := bufio.NewScanner(pout)
		for scanner.Scan() {
			t.Log(scanner.Text())
		}
	}()

	ch := make(chan error, 1)

	go func() {
		err = cmd.Start()
		if err != nil {
			ch <- errors.Wrapf(err, "Failed to start process")
			return
		}
		err = cmd.Wait()
		if err != nil {
			ch <- errors.Wrapf(err, "Failed to wait")
			return
		}

		ch <- nil
		close(ch)
	}()

	select {
	case err := <-ch:
		if err != nil {
			t.Errorf("Command failed:\n%s", err.Error())
		}
	case <-ctx.Done():
		t.Errorf("Command timed out")
	}
}

func validateManifest(t *testing.T, chartfile string) {
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
}
