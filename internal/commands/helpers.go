package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/sachinkaru123/giteasy/internal/config"
	"github.com/sachinkaru123/giteasy/internal/git"
	"github.com/sachinkaru123/giteasy/internal/ui"
)

// requireRepo exits with a friendly message if git isn't installed or
// the current directory isn't a git repository.
func requireRepo() {
	if !git.IsInstalled() {
		ui.Error("git was not found on your PATH. Install git and try again.")
		os.Exit(1)
	}
	if !git.IsRepo() {
		ui.Error("This directory doesn't look like a git repository.")
		os.Exit(1)
	}
}

// pickBranch shows "primary / secondary / other" and, if "other" is chosen,
// lets the user pick from every local+remote branch. Returns the chosen name.
func pickBranch(promptTitle string, cfg *config.Config, exclude ...string) string {
	primary := cfg.DefaultBranches.Primary
	secondary := cfg.DefaultBranches.Secondary

	options := []string{primary, secondary, "other →"}
	idx, choice := ui.Select(promptTitle, options)
	if idx != len(options)-1 {
		return choice
	}

	all, err := git.ListAllBranchNames()
	if err != nil {
		ui.Error("Could not list branches: %v", err)
		os.Exit(1)
	}
	var filtered []string
	for _, b := range all {
		skip := false
		for _, ex := range exclude {
			if b == ex {
				skip = true
			}
		}
		if !skip {
			filtered = append(filtered, b)
		}
	}
	if len(filtered) == 0 {
		ui.Error("No other branches found.")
		os.Exit(1)
	}
	_, picked := ui.Select("Choose a branch:", filtered)
	return picked
}

// resolveDefaultBranches ensures cfg.DefaultBranches.Primary actually exists
// (main vs master can vary per repo); auto-detects once and persists it.
func resolveDefaultBranches(cfg *config.Config) {
	if git.BranchExists(cfg.DefaultBranches.Primary) || git.RemoteBranchExists(cfg.DefaultBranches.Primary) {
		return
	}
	detected, err := git.RemoteHeadBranch()
	if err == nil && detected != "" {
		cfg.DefaultBranches.Primary = detected
		_ = cfg.Save()
	}
}

// resolvePrefix computes the branch-name prefix per settings.
func resolvePrefix(cfg *config.Config) string {
	switch cfg.Branch.PrefixMode {
	case config.PrefixCustom:
		return cfg.Branch.CustomPrefix
	case config.PrefixGitUserFirstName:
		name := git.ConfigUserName()
		if name == "" {
			return ""
		}
		first := strings.Fields(name)
		if len(first) == 0 {
			return ""
		}
		return strings.ToLower(first[0]) + "/"
	default:
		return ""
	}
}

// mustNoError prints a formatted error and exits(1) if err != nil.
func mustNoError(err error, context string) {
	if err != nil {
		ui.Error("%s: %v", context, err)
		os.Exit(1)
	}
}

func printRebaseConflictHelp() {
	fmt.Println()
	ui.Warn("Rebase stopped due to a conflict.")
	fmt.Println("  Resolve the conflicting files, then run:")
	fmt.Println("    git add <files>")
	fmt.Println("    git rebase --continue")
	fmt.Println("  Or to abort and go back to where you started:")
	fmt.Println("    git rebase --abort")
}

func printMergeConflictHelp() {
	fmt.Println()
	ui.Warn("Merge stopped due to a conflict.")
	fmt.Println("  Resolve the conflicting files, then run:")
	fmt.Println("    git add <files>")
	fmt.Println("    git commit")
	fmt.Println("  Or to abort and go back to where you started:")
	fmt.Println("    git merge --abort")
}
