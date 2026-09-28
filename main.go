// Command giteasy is a small, dependency-free CLI that wraps common
// git workflows (branch creation with auto-sync, tag suggestion,
// squash, rebase-with-fetch, merge with guardrails, and reset) behind
// a friendly interactive menu, while still supporting direct
// subcommands for scripting / muscle memory.
package main

import (
	"fmt"
	"os"

	"github.com/sachinkaru123/giteasy/internal/commands"
)

// version is set at build time via -ldflags "-X main.version=..." by
// GoReleaser. Defaults to "dev" for local `go build`/`go run`.
var version = "dev"

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		commands.RunMenu()
		return
	}

	switch args[0] {
	case "new":
		if len(args) < 2 {
			printNewUsage()
			os.Exit(1)
		}
		switch args[1] {
		case "branch":
			commands.NewBranch()
		case "tag":
			commands.NewTag()
		default:
			printNewUsage()
			os.Exit(1)
		}
	case "push":
		commands.Push()
	case "sync":
		commands.Sync()
	case "squash":
		commands.Squash()
	case "rebase":
		commands.Rebase()
	case "merge":
		commands.Merge()
	case "reset":
		commands.Reset()
	case "settings":
		commands.Settings()
	case "-v", "--version", "version":
		fmt.Println("giteasy version " + version)
	case "-h", "--help", "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n\n", args[0])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Print(`giteasy - a friendly git workflow helper

Usage:
  giteasy                 Open the interactive menu
  giteasy new branch      Create a branch (local + origin, auto-synced)
  giteasy new tag         Suggest and create the next tag
  giteasy push            Push the current branch
  giteasy sync            Fetch from origin and pull the current branch
  giteasy squash          Squash commits into one
  giteasy rebase          Rebase onto another branch (fetches it first)
  giteasy merge           Merge one branch into another
  giteasy reset           Reset a branch (soft/mixed/hard)
  giteasy settings        Configure giteasy
  giteasy help            Show this message
  giteasy version         Show version
`)
}

func printNewUsage() {
	fmt.Println("Usage: giteasy new <branch|tag>")
}
