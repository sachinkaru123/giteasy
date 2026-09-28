package commands

import (
	"github.com/yourname/giteasy/internal/config"
	"github.com/yourname/giteasy/internal/git"
	"github.com/yourname/giteasy/internal/ui"
)

// Merge merges a chosen source branch into a chosen target branch.
// If the target is the repo's primary (main) branch and the source is
// *not* the secondary (dev) branch, it asks for confirmation first
// (this guard is on by default, toggle it off in Settings).
func Merge() {
	requireRepo()
	ui.Header("Merge")
	if !git.IsWorkingTreeClean() {
		ui.Error("Working tree has uncommitted changes. Commit or stash them first.")
		return
	}

	cfg, err := config.Load()
	mustNoError(err, "loading config")
	resolveDefaultBranches(cfg)

	target := pickBranch("Merge INTO which branch (target):", cfg)
	source := pickBranch("Merge FROM which branch (source):", cfg, target)

	if cfg.Merge.ConfirmNonDevToMain &&
		target == cfg.DefaultBranches.Primary &&
		source != cfg.DefaultBranches.Secondary {
		warned := ui.Confirm(
			"⚠ You're merging \""+source+"\" directly into \""+target+"\" (not via "+cfg.DefaultBranches.Secondary+"). Continue?",
			false,
		)
		if !warned {
			ui.Info("Merge cancelled.")
			return
		}
	}

	ui.Info("Fetching from origin ...")
	mustNoError(git.FetchAll(), "fetching")

	ui.Info("Checking out %q and updating it ...", target)
	mustNoError(git.CheckoutBranch(target), "checking out "+target)
	mustNoError(git.PullBranch(target), "pulling "+target)

	mergeSourceRef := source
	if git.BranchExists(source) {
		mustNoError(git.PullBranch(source), "updating "+source)
	} else if git.RemoteBranchExists(source) {
		mergeSourceRef = "origin/" + source
	}

	ui.Info("Merging %q into %q ...", source, target)
	conflict, err := git.MergeBranch(mergeSourceRef)
	if conflict {
		printMergeConflictHelp()
		return
	}
	if err != nil {
		ui.Error("Merge failed: %v", err)
		return
	}

	if err := git.Push(target); err != nil {
		ui.Error("Merged locally, but push failed: %v", err)
		return
	}
	ui.Success("Merged %q into %q and pushed.", source, target)

	deleteBranch := false
	switch cfg.Merge.DeleteAfterMerge {
	case config.DeleteAlways:
		deleteBranch = true
	case config.DeleteNever:
		deleteBranch = false
	case config.DeleteAsk:
		deleteBranch = ui.Confirm("Delete \""+source+"\" branch after merge?", false)
	}

	if deleteBranch {
		if git.BranchExists(source) {
			mustNoError(git.DeleteLocalBranch(source), "deleting local "+source)
		}
		if git.RemoteBranchExists(source) {
			mustNoError(git.DeleteRemoteBranch(source), "deleting origin/"+source)
		}
		ui.Success("Deleted branch %q locally and on origin.", source)
	}
}
