//go:build !windows

package ui

import (
	"os"
	"os/exec"
	"strings"
)

func stty(args ...string) (string, error) {
	cmd := exec.Command("stty", args...)
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}

// makeRaw puts the terminal into single-keypress mode (no line buffering,
// no echo, Ctrl+C delivered as a byte) and returns a function that restores
// the previous state. Output post-processing stays on so "\n" still works.
func makeRaw() (func(), error) {
	state, err := stty("-g")
	if err != nil {
		return nil, err
	}
	if _, err := stty("-icanon", "-echo", "-isig", "min", "1", "time", "0"); err != nil {
		return nil, err
	}
	return func() { _, _ = stty(state) }, nil
}

// enableVT reports whether ANSI escape sequences are usable (always, on unix).
func enableVT() bool { return true }
