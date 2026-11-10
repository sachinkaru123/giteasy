package commands

import (
	"testing"

	"github.com/sachinkaru123/giteasy/internal/config"
)

func TestIsProtectedBranch(t *testing.T) {
	cfg := &config.Config{}
	cfg.DefaultBranches.Primary = "main"
	cfg.DefaultBranches.Secondary = "dev"
	for name, want := range map[string]bool{"main": true, "dev": true, "feat/x": false} {
		if got := isProtectedBranch(name, cfg); got != want {
			t.Errorf("isProtectedBranch(%q) = %v, want %v", name, got, want)
		}
	}
}
