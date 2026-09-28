package commands

import (
	"github.com/yourname/giteasy/internal/config"
	"github.com/yourname/giteasy/internal/git"
	"github.com/yourname/giteasy/internal/ui"
)

// Rebase rebases the current branch onto a chosen target. Per spec: before
// rebasing, it fetches and fast-forwards the *target* branch so the rebase
// is always against the latest remote state, not a stale local copy.
func Rebase() {
	requireRepo()
	if !git.IsWorkingTreeClean() {
		ui.Error("Working tree has uncommitted changes. Commit or stash them first.")
		return
	}

	cfg, err := config.Load()
	mustNoError(err, "loading config")
	resolveDefaultBranches(cfg)

	current, err := git.CurrentBranch()
	mustNoError(err, "getting current branch")

	target := pickBranch("Rebase current branch onto:", cfg, current)

	ui.Info("Fetching from origin ...")
	mustNoError(git.FetchAll(), "fetching")

	// Bring the target branch up to date before rebasing onto it.
	if git.BranchExists(target) {
		ui.Info("Checking out %q to update it ...", target)
		mustNoError(git.CheckoutBranch(target), "checking out "+target)
		if err := git.PullBranch(target); err != nil {
			ui.Error("Could not update %q: %v", target, err)
			_ = git.CheckoutBranch(current)
			return
		}
		mustNoError(git.CheckoutBranch(current), "returning to "+current)
	} else if git.RemoteBranchExists(target) {
		ui.Info("%q only exists on origin, using origin/%s as rebase target.", target, target)
		target = "origin/" + target
	} else {
		ui.Error("Branch %q not found locally or on origin.", target)
		return
	}

	ui.Info("Rebasing %q onto %q ...", current, target)
	conflict, err := git.RebaseOnto(target)
	if conflict {
		printRebaseConflictHelp()
		return
	}
	if err != nil {
		ui.Error("Rebase failed: %v", err)
		return
	}
	ui.Success("Rebased %q onto %q.", current, target)

	if git.HasUpstream(current) {
		if ui.Confirm("Force-push the rebased branch to origin?", true) {
			if err := git.ForcePush(current); err != nil {
				ui.Error("Force-push failed: %v", err)
				return
			}
			ui.Success("Pushed rebased branch to origin.")
		}
	}
}
