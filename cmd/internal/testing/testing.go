package testing

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/pkg/errors"
)

func Run(t *testing.T, args ...string) (string, int) {
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

	// Set up a chain of pipes.  Out is a buffer to return STDOUT.
	// The command's STDERR will only go to the bufio scanner for dumping to the
	// test log.  The command's STDOUT will be duplicated to write both to the
	// test log and to an in-memory buffer, to be returned as a string.
	var out bytes.Buffer
	pout, pin := io.Pipe()
	min := io.MultiWriter(pin, &out)

	cmd.Stdout = min
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
		if _, ok := err.(*exec.ExitError); err != nil && !ok {
			// We have an err, and it's not an ExitError (trapping a non-zero exit
			// code).
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

	return out.String(), cmd.ProcessState.ExitCode()
}
