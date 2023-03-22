package neotest

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
)

// LoadJSON unmarshals JSON for a test
func LoadJSON(t *testing.T, path string, dst interface{}) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s", path)
	}
	if err := json.Unmarshal(bytes, dst); err != nil {
		t.Fatalf("failed to unmarshal %s", path)
	}
}

// CloneTestData clones the test data into a temporary directory and changes the
// working directory to the temporary directory. The function returned should be
// run as a deferred to restore the working directory and remove the files.
func CloneTestData(t *testing.T) (string, func()) {
	dir, err := ioutil.TempDir("", "testdata")
	if err != nil {
		t.Fatal(err)
	}

	// Makes me sad.
	cmd := exec.Command("cp", "-r", ".", dir)
	cmd.Dir = "testdata"
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}

	curdir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	return dir, func() {
		if err := os.Chdir(curdir); err != nil {
			t.Fatal(err)
		}
		if err := os.RemoveAll(dir); err != nil {
			t.Errorf("error removing temp dir: %s", err)
		}
	}
}

// Must calls fatal in the given test if err is non-nil
func Must(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

// Dump i to stdout using JSON
func Dump(t *testing.T, i interface{}) {
	if data, err := json.MarshalIndent(i, "", "  "); err != nil {
		if t != nil {
			t.Fatal(err)
		} else {
			panic(err)
		}
	} else {
		println(string(data))
	}
}
