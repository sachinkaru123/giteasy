package commands

import (
	"strings"

	"github.com/sachinkaru123/giteasy/internal/git"
	"github.com/sachinkaru123/giteasy/internal/ui"
)

// NewTag shows the latest tag, suggests the next patch version pre-filled
// and editable, and on Enter creates + pushes the tag.
func NewTag() {
	requireRepo()
	ui.Header("New Tag")

	latest, err := git.LatestTag()
	mustNoError(err, "reading tags")

	if latest != "" {
		ui.Info("Latest tag: %s", latest)
	} else {
		ui.Info("No existing tags found.")
	}

	suggestion := git.SuggestNextPatch(latest)
	name := ui.Input("New tag", suggestion)
	if strings.TrimSpace(name) == "" {
		ui.Error("Tag name cannot be empty.")
		return
	}

	if err := git.CreateTag(name); err != nil {
		ui.Error("Failed to create tag: %v", err)
		return
	}
	if err := git.PushTag(name); err != nil {
		ui.Error("Tag created locally but push failed: %v", err)
		return
	}
	ui.Success("Tag %q created and pushed to origin.", name)
}
