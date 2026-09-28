package commands

import "github.com/sachinkaru123/giteasy/internal/ui"

// RunMenu shows the top-level interactive menu (bare `giteasy`).
func RunMenu() {
	options := []string{
		"New Branch",
		"New Tag",
		"Push",
		"Sync (fetch/pull)",
		"Squash Commits",
		"Rebase",
		"Merge",
		"Reset Branch",
		"Settings",
		"Exit",
	}
	ui.Header()
	for {
		idx, _ := ui.Select("What do you want to do?", options)
		switch idx {
		case 0:
			NewBranch()
		case 1:
			NewTag()
		case 2:
			Push()
		case 3:
			Sync()
		case 4:
			Squash()
		case 5:
			Rebase()
		case 6:
			Merge()
		case 7:
			Reset()
		case 8:
			Settings()
		case 9:
			return
		}
	}
}
