package mcov_test

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var env []string

func init() {
	testdata, err := filepath.Abs("testdata")
	if err != nil {
		log.Fatalf("filepath.Abs()=%s", err)
	}

	gopath := testdata + string(os.PathListSeparator) + os.Getenv("GOPATH")

	env = os.Environ()

	for i, s := range env {
		if strings.HasPrefix(s, "GOPATH=") {
			env[i] = "GOPATH=" + gopath
			return
		}
	}

	env = append(env, "GOPATH="+gopath)
}

func TestExample(t *testing.T) {
	coverprofile := filepath.FromSlash("testdata/src/example/example.cover")
	defer os.Remove(coverprofile)

	_, err := run("go", "test", "-coverprofile", coverprofile, "example")
	if err != nil {
		t.Fatalf("go test=%s", err)
	}

	prof, err := run("go", "tool", "cover", "-func", coverprofile)
	if err != nil {
		t.Fatalf("go tool cover=%s", err)
	}

	total := "-"

	if i := strings.LastIndex(prof, "\t"); i != -1 {
		total = prof[i+1:]
	}

	if want := "100.0%"; total != want {
		t.Fatalf("got %q, want %q", total, want)
	}
}

func run(cmd string, args ...string) (string, error) {
	var buf bytes.Buffer

	c := exec.Command(cmd, args...)
	c.Env = env
	c.Stdout = &buf
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		return "", err
	}

	return strings.TrimSpace(buf.String()), nil
}
