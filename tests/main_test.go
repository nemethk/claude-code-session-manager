package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

var binary string

func TestMain(m *testing.M) {
	tmp, err := os.MkdirTemp("", "ccsm-e2e-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "setup: %v\n", err)
		os.Exit(1)
	}
	binary = filepath.Join(tmp, "ccsm")

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		fmt.Fprintf(os.Stderr, "setup: runtime.Caller failed\n")
		os.RemoveAll(tmp)
		os.Exit(1)
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), ".."))

	out, err := exec.Command("go", "build", "-o", binary, root).CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "build failed: %v\n%s\n", err, out)
		os.RemoveAll(tmp)
		os.Exit(1)
	}

	code := m.Run()
	os.RemoveAll(tmp)
	os.Exit(code)
}
