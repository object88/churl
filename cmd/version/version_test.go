package version

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/jmespath/go-jmespath"
	ctesting "github.com/object88/churl/cmd/internal/testing"
)

func Test_Cmd_Version(t *testing.T) {
	expectedSha := os.Getenv("TEST_SHA")
	expectedVersion := os.Getenv("TEST_VERSION")

	if expectedSha == "" || expectedVersion == "" {
		t.Skipf("Test requires both $TEST_SHA and $TEST_VERSION")
	}

	out := ctesting.Run(t, "version", "--output", "json")
	if out == "Churl version unset\n" {
		t.Errorf("Version is not set: '%s'", out)
	}

	var data interface{}
	json.NewDecoder(strings.NewReader(out)).Decode(&data)

	test(t, expectedSha, data, "sha")
	test(t, expectedVersion, data, "version")
}

func test(t *testing.T, expected string, data interface{}, query string) {
	jmes, err := jmespath.Compile(query)
	if err != nil {
		t.Fatalf("Failed to compile query '%s':\n%s", query, err.Error())
	}

	if actual, err := jmes.Search(data); err != nil {
		t.Fatalf("Failed to execute jmes query with query '%s':\n%s", query, err.Error())
	} else if actual != expected {
		t.Errorf("Got wrong version: expected '%s', actual '%s'", expected, actual)
	}

}
