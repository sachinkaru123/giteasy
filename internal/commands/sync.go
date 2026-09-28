package commands

import (
	"github.com/yourname/giteasy/internal/git"
	"github.com/yourname/giteasy/internal/ui"
)

// Sync fetches everything from origin (pruning deleted remote branches)
// and fast-forwards the current branch if it has an upstream.
func Sync() {
	requireRepo()
	ui.Info("Fetching from origin ...")
	if err := git.FetchAll(); err != nil {
		ui.Error("Fetch failed: %v", err)
		return
	}

	branch, err := git.CurrentBranch()
	mustNoError(err, "getting current branch")

	if !git.HasUpstream(branch) {
		ui.Success("Fetched. %q has no upstream yet, nothing to pull.", branch)
		return
	}

	ui.Info("Pulling %q up to date ...", branch)
	if err := git.PullBranch(branch); err != nil {
		ui.Error("Pull failed: %v", err)
		return
	}
	ui.Success("%q is up to date with origin.", branch)
}
