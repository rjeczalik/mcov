package example_test

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"example"

	"github.com/rjeczalik/mcov"
)

func die(v ...interface{}) {
	fmt.Fprintln(os.Stderr, v...)
	os.Exit(1)
}

func TestHelper(t *testing.T) {
	if os.Getenv("TEST_HELPER") != "1" {
		return
	}
	defer os.Stdout.Close()

	args := os.Args

	for i, arg := range args {
		if arg == "--" {
			args = args[i+1:]
			break
		}
	}

	if len(args) != 2 {
		die("invalid number of arguments:", len(args))
	}

	switch args[0] {
	case "echo":
		fmt.Println(example.Echo(args[1]))
	case "reverse":
		fmt.Println(example.Reverse(args[1]))
	default:
		die("unrecognized argument:", args[0])
	}
}

func TestMain(m *testing.M) {
	flag.Parse()
	code := m.Run()

	if os.Getenv("TEST_HELPER") == "" {
		if err := mcov.SubprofileMerge(""); err != nil {
			die(err)
		}
	}

	os.Exit(code)
}

func call(fun, arg string) (string, error) {
	var buf bytes.Buffer

	cmd := exec.Command(os.Args[0], append(mcov.SubprofileFlags([]string{"-test.run=TestHelper"}), "--", fun, arg)...)
	cmd.Env = append(os.Environ(), "TEST_HELPER=1")
	cmd.Stdout = &buf
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", err
	}

	return strings.TrimSpace(buf.String()), nil
}
