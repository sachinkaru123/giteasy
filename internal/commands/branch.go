package commands

import (
	"strings"

	"github.com/yourname/giteasy/internal/config"
	"github.com/yourname/giteasy/internal/git"
	"github.com/yourname/giteasy/internal/ui"
)

// NewBranch runs the "create branch" flow: pick a base branch, name the
// new branch (with prefix pre-filled per settings, editable), create it
// locally, check it out, and push it to origin so local/remote stay in sync.
func NewBranch() {
	requireRepo()
	cfg, err := config.Load()
	mustNoError(err, "loading config")
	resolveDefaultBranches(cfg)

	base := pickBranch("Create branch from:", cfg)

	prefix := resolvePrefix(cfg)
	name := ui.Input("Branch name", prefix)
	if strings.TrimSpace(name) == "" || strings.TrimSpace(name) == strings.TrimSpace(prefix) {
		ui.Error("Branch name cannot be empty.")
		return
	}

	ui.Info("Creating %q from %q ...", name, base)
	if err := git.CreateBranch(name, base); err != nil {
		ui.Error("Failed to create branch: %v", err)
		return
	}
	ui.Success("Branch %q created locally and pushed to origin.", name)
}
