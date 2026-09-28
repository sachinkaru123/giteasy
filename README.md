# giteasy

A friendly CLI that wraps common git workflows — branch creation with
auto-sync, tag suggestion, squashing, safe rebasing, guarded merging, and
branch resets — behind a simple interactive menu. Zero dependencies, one
static binary.

```
$ giteasy

What do you want to do?
  1) New Branch
  2) New Tag
  3) Push
  4) Sync (fetch/pull)
  5) Squash Commits
  6) Rebase
  7) Merge
  8) Reset Branch
  9) Settings
  10) Exit
Enter choice number:
```

## Install

**macOS / Linux**
```bash
curl -sSL https://raw.githubusercontent.com/sachinkaru123/giteasy/main/install.sh | bash
```

**Windows (PowerShell)**
```powershell
irm https://raw.githubusercontent.com/sachinkaru123/giteasy/main/install.ps1 | iex
```

**Already have Go?**
```bash
go install github.com/sachinkaru123/giteasy@latest
```

No other dependencies are required on your machine — the installers just
download a prebuilt binary. `giteasy` shells out to your existing `git`
installation, so your normal auth/SSH setup just works.

## Commands

Running `giteasy` with no arguments opens the interactive menu. Every
action is also available as a direct subcommand:

| Command | What it does |
|---|---|
| `giteasy new branch` | Pick a base (main/dev/other), name the branch (prefix pre-filled per settings), create it locally, check it out, and push it to origin — local and remote stay in sync automatically. |
| `giteasy new tag` | Shows the latest tag, suggests the next patch version (e.g. `v1.0.1` → `v1.0.2`), editable — press Enter to accept and it creates + pushes the tag. |
| `giteasy push` | Pushes the current branch, setting upstream if it's missing. |
| `giteasy sync` | Fetches everything from origin (pruning deleted branches) and fast-forwards the current branch. |
| `giteasy squash` | Squash the last N commits, or everything since diverging from main/dev, into one commit with a message you provide. Offers to force-push after. |
| `giteasy rebase` | Pick a target branch. Fetches and fast-forwards *that branch* first, then rebases your current branch onto it — so you're never rebasing onto stale history. Gives clear `--continue`/`--abort` guidance on conflicts. |
| `giteasy merge` | Pick a target and a source branch, updates both from origin, merges, and pushes. If you merge something other than `dev` directly into `main`, it asks for confirmation first (toggle this in Settings). Offers to delete the source branch after, per your settings. |
| `giteasy reset` | Pick a branch, a reset mode (soft/mixed/hard), and a target (origin, a commit hash, or N commits back). Hard resets require confirmation. Offers a force-push afterward if the branch is already on origin. |
| `giteasy settings` | Configure branch-name prefixes, merge-deletion behavior, the main→merge confirmation guard, and which branches count as "primary"/"secondary" for your repo. |

## Settings

Settings are stored in `~/.giteasy/config.json`:

```json
{
  "branch": {
    "prefixMode": "custom",
    "customPrefix": "feat/"
  },
  "merge": {
    "deleteAfterMerge": "ask",
    "confirmNonDevToMain": true
  },
  "defaultBranches": {
    "primary": "main",
    "secondary": "dev"
  }
}
```

- **branch.prefixMode**: `gitUserFirstName` (derives a prefix from your
  `git config user.name`), `custom`, or `none`.
- **merge.deleteAfterMerge**: `always`, `never`, or `ask` — whether the
  source branch is deleted (locally and on origin) after a successful merge.
- **merge.confirmNonDevToMain**: when `true` (default), merging any branch
  other than your "secondary" branch directly into "primary" asks for
  confirmation first.
- **defaultBranches**: which branches `giteasy` treats as primary/secondary
  in the branch pickers. Primary is auto-detected (`main` vs `master`) on
  first use if the configured one doesn't exist.

## Building from source

```bash
git clone https://github.com/sachinkaru123/giteasy.git
cd giteasy
go build -o giteasy .
```

No external Go modules are required — this repo has zero third-party
dependencies, just the standard library.

## Releasing (for maintainers)

Releases are built automatically by [GoReleaser](https://goreleaser.com/)
via GitHub Actions whenever a version tag is pushed:

```bash
git tag v1.0.0
git push origin v1.0.0
```

This builds binaries for Linux/macOS/Windows (amd64 + arm64) and attaches
them to a GitHub Release, which `install.sh` / `install.ps1` then pull from.

## License

MIT — see [LICENSE](./LICENSE).
