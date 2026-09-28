package commands

import (
	"github.com/yourname/giteasy/internal/git"
	"github.com/yourname/giteasy/internal/ui"
)

// Reset lets the user reset a branch: choose which branch, choose the
// reset mode (soft/mixed/hard), and choose the target point to reset to.
func Reset() {
	requireRepo()
	ui.Header("Reset")

	current, err := git.CurrentBranch()
	mustNoError(err, "getting current branch")

	branches, err := git.ListAllBranchNames()
	mustNoError(err, "listing branches")

	options := []string{"current (" + current + ")"}
	options = append(options, branches...)
	idx, choice := ui.Select("Which branch do you want to reset?", options)
	branch := choice
	if idx == 0 {
		branch = current
	}

	if branch != current {
		if !git.IsWorkingTreeClean() {
			ui.Error("Working tree has uncommitted changes. Commit or stash them before switching branches.")
			return
		}
		mustNoError(git.CheckoutBranch(branch), "checking out "+branch)
	}

	modeOptions := []string{
		"Soft  (keep changes staged)",
		"Mixed (keep changes unstaged) — git default",
		"Hard  (discard all changes) ⚠",
	}
	modeIdx, _ := ui.Select("Reset mode:", modeOptions)
	var mode git.ResetMode
	switch modeIdx {
	case 0:
		mode = git.ResetSoft
	case 1:
		mode = git.ResetMixed
	case 2:
		mode = git.ResetHard
	}

	targetOptions := []string{
		"origin/" + branch + " (discard local commits, match remote)",
		"Specific commit (enter hash)",
		"N commits back",
	}
	targetIdx, _ := ui.Select("Reset to:", targetOptions)

	var target string
	switch targetIdx {
	case 0:
		if !git.RemoteBranchExists(branch) {
			ui.Error("origin/%s doesn't exist, nothing to reset to.", branch)
			return
		}
		target = "origin/" + branch
	case 1:
		target = ui.Input("Commit hash", "")
		if target == "" {
			ui.Error("Commit hash cannot be empty.")
			return
		}
	case 2:
		n := ui.Input("How many commits back", "1")
		target = "HEAD~" + n
	}

	if mode == git.ResetHard {
		confirmed := ui.Confirm("This will PERMANENTLY discard local changes on \""+branch+"\". Continue?", false)
		if !confirmed {
			ui.Info("Reset cancelled.")
			return
		}
	}

	ui.Info("Resetting %q (%s) to %s ...", branch, mode, target)
	if err := git.ResetTo(mode, target); err != nil {
		ui.Error("Reset failed: %v", err)
		return
	}
	ui.Success("Branch %q reset to %s.", branch, target)

	if git.HasUpstream(branch) {
		if ui.Confirm("This branch is pushed to origin. Force-push the reset?", false) {
			if err := git.ForcePush(branch); err != nil {
				ui.Error("Force-push failed: %v", err)
				return
			}
			ui.Success("Force-pushed reset to origin.")
		}
	}
}
