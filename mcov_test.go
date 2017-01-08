package mcov_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var gopath = "testdata" + string(os.PathListSeparator) + os.Getenv("GOPATH")

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
	c.Env = append(os.Environ(), "GOPATH="+gopath)
	c.Stdout = &buf
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		return "", err
	}

	return strings.TrimSpace(buf.String()), nil
}
