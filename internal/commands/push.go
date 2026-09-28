package commands

import (
	"github.com/yourname/giteasy/internal/git"
	"github.com/yourname/giteasy/internal/ui"
)

// Push pushes the current branch, setting upstream automatically if missing.
func Push() {
	requireRepo()
	branch, err := git.CurrentBranch()
	mustNoError(err, "getting current branch")

	ui.Info("Pushing %q ...", branch)
	if err := git.Push(branch); err != nil {
		ui.Error("Push failed: %v", err)
		return
	}
	ui.Success("Pushed %q to origin.", branch)
}
