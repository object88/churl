package e2e

import (
	"testing"

	ctesting "github.com/object88/churl/internal/testing"
)

func findBinary(t *testing.T, binary string) *ctesting.Runnable {
	bin, ok, err := ctesting.NewRunnable(t, binary)
	if err != nil {
		t.Fatalf("Internal error finding '%s' binary:\n%s", binary, err.Error())
	}
	if !ok {
		t.Fatalf("Failed to find '%s' binary, cannot run tests", binary)
	}
	return bin
}
