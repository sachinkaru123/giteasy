package commands

import (
	"strings"

	"github.com/yourname/giteasy/internal/config"
	"github.com/yourname/giteasy/internal/git"
	"github.com/yourname/giteasy/internal/ui"
)

// Settings shows the settings menu and loops until the user backs out.
func Settings() {
	cfg, err := config.Load()
	mustNoError(err, "loading config")

	for {
		options := []string{
			"Branch prefix",
			"Merge: delete branch after merge",
			"Merge: confirm non-dev → main merges",
			"Default branches (primary/secondary)",
			"Back",
		}
		idx, _ := ui.Select("Settings", options)
		switch idx {
		case 0:
			settingsBranchPrefix(cfg)
		case 1:
			settingsDeleteAfterMerge(cfg)
		case 2:
			settingsConfirmNonDevToMain(cfg)
		case 3:
			settingsDefaultBranches(cfg)
		case 4:
			return
		}
	}
}

func settingsBranchPrefix(cfg *config.Config) {
	options := []string{
		"Use git user.name (first name)",
		"Custom prefix",
		"None",
	}
	idx, _ := ui.Select("Branch prefix mode:", options)
	switch idx {
	case 0:
		suggestion := ""
		if name := git.ConfigUserName(); name != "" {
			suggestion = name
		}
		confirmed := ui.Input("Confirm/edit the name to derive the prefix from", suggestion)
		cfg.Branch.PrefixMode = config.PrefixGitUserFirstName
		// Store as custom under the hood if user edited it into something
		// that isn't their literal git user.name, so the choice is precise.
		if confirmed != suggestion {
			cfg.Branch.PrefixMode = config.PrefixCustom
			first := confirmed
			cfg.Branch.CustomPrefix = firstWordLowerSlash(first)
		}
	case 1:
		prefix := ui.Input("Custom prefix (e.g. \"feat/\")", cfg.Branch.CustomPrefix)
		cfg.Branch.PrefixMode = config.PrefixCustom
		cfg.Branch.CustomPrefix = prefix
	case 2:
		cfg.Branch.PrefixMode = config.PrefixNone
	}
	mustNoError(cfg.Save(), "saving config")
	ui.Success("Branch prefix setting saved.")
}

func settingsDeleteAfterMerge(cfg *config.Config) {
	options := []string{"Always", "Never", "Ask each time"}
	idx, _ := ui.Select("Delete source branch after merge:", options)
	switch idx {
	case 0:
		cfg.Merge.DeleteAfterMerge = config.DeleteAlways
	case 1:
		cfg.Merge.DeleteAfterMerge = config.DeleteNever
	case 2:
		cfg.Merge.DeleteAfterMerge = config.DeleteAsk
	}
	mustNoError(cfg.Save(), "saving config")
	ui.Success("Saved.")
}

func settingsConfirmNonDevToMain(cfg *config.Config) {
	options := []string{"On (recommended)", "Off"}
	idx, _ := ui.Select("Confirm before merging a non-dev branch into main:", options)
	cfg.Merge.ConfirmNonDevToMain = idx == 0
	mustNoError(cfg.Save(), "saving config")
	ui.Success("Saved.")
}

func settingsDefaultBranches(cfg *config.Config) {
	primary := ui.Input("Primary branch name (main/master)", cfg.DefaultBranches.Primary)
	secondary := ui.Input("Secondary branch name (dev)", cfg.DefaultBranches.Secondary)
	cfg.DefaultBranches.Primary = primary
	cfg.DefaultBranches.Secondary = secondary
	mustNoError(cfg.Save(), "saving config")
	ui.Success("Saved.")
}

func firstWordLowerSlash(s string) string {
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return ""
	}
	return strings.ToLower(fields[0]) + "/"
}
