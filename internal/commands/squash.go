package commands

import (
	"strconv"
	"strings"

	"github.com/yourname/giteasy/internal/config"
	"github.com/yourname/giteasy/internal/git"
	"github.com/yourname/giteasy/internal/ui"
)

// Squash lets the user squash either the last N commits, or every commit
// since the current branch diverged from main/dev, into a single commit.
func Squash() {
	requireRepo()
	cfg, err := config.Load()
	mustNoError(err, "loading config")
	resolveDefaultBranches(cfg)

	options := []string{
		"Last N commits",
		"Since diverging from " + cfg.DefaultBranches.Primary,
		"Since diverging from " + cfg.DefaultBranches.Secondary,
	}
	idx, _ := ui.Select("Squash last N commits, or since branching?", options)

	var n int
	switch idx {
	case 0:
		raw := ui.Input("How many commits to squash", "2")
		n, err = strconv.Atoi(strings.TrimSpace(raw))
		if err != nil || n < 2 {
			ui.Error("Enter a number 2 or greater.")
			return
		}
	case 1, 2:
		base := cfg.DefaultBranches.Primary
		if idx == 2 {
			base = cfg.DefaultBranches.Secondary
		}
		count, err := git.CommitsAhead(base)
		mustNoError(err, "counting commits")
		if count < 2 {
			ui.Warn("Only %d commit(s) ahead of %q, nothing to squash.", count, base)
			return
		}
		n = count
		ui.Info("Found %d commits ahead of %q.", n, base)
	}

	message := ui.Input("Commit message", "")
	if strings.TrimSpace(message) == "" {
		ui.Error("Commit message cannot be empty.")
		return
	}

	if err := git.SquashLastN(n, message); err != nil {
		ui.Error("Squash failed: %v", err)
		return
	}
	ui.Success("Squashed %d commits into one.", n)

	branch, err := git.CurrentBranch()
	if err == nil && git.HasUpstream(branch) {
		if ui.Confirm("Force-push the squashed history to origin?", true) {
			if err := git.ForcePush(branch); err != nil {
				ui.Error("Force-push failed: %v", err)
				return
			}
			ui.Success("Pushed squashed commit to origin.")
		}
	}
}
