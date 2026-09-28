// Package git provides thin wrappers around the system git binary.
// giteasy intentionally shells out to git rather than reimplementing
// git internals, so it inherits the user's existing auth, SSH keys,
// and .gitconfig for free.
package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Run executes `git <args...>` and returns trimmed stdout.
// On failure it returns an error whose message includes both stdout and
// stderr for context - git puts some important messages (e.g. merge
// conflict notices) on stdout rather than stderr.
func Run(args ...string) (string, error) {
	out, errOut, err := runCapture(args...)
	if err != nil {
		msg := strings.TrimSpace(errOut)
		if msg == "" {
			msg = strings.TrimSpace(out)
		}
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), msg)
	}
	return strings.TrimSpace(out), nil
}

// runCapture runs git and returns stdout, stderr, and the exec error
// separately, so callers that need to inspect both streams (e.g. to
// detect a merge/rebase conflict, which git reports on stdout) can do so.
func runCapture(args ...string) (stdout, stderr string, err error) {
	cmd := exec.Command("git", args...)
	var out, errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	err = cmd.Run()
	return out.String(), errBuf.String(), err
}

// IsInstalled reports whether the git binary is available on PATH.
func IsInstalled() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

// IsRepo reports whether the current directory is inside a git work tree.
func IsRepo() bool {
	out, err := Run("rev-parse", "--is-inside-work-tree")
	return err == nil && out == "true"
}

// CurrentBranch returns the name of the currently checked out branch.
func CurrentBranch() (string, error) {
	return Run("rev-parse", "--abbrev-ref", "HEAD")
}

// RemoteHeadBranch asks origin which branch HEAD points to (e.g. "main").
// Falls back to checking for local "main" then "master" if origin/HEAD
// isn't set (common on fresh clones or local-only repos).
func RemoteHeadBranch() (string, error) {
	out, err := Run("symbolic-ref", "refs/remotes/origin/HEAD")
	if err == nil {
		parts := strings.Split(out, "/")
		return parts[len(parts)-1], nil
	}
	for _, candidate := range []string{"main", "master"} {
		if BranchExists(candidate) {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("could not determine default branch (no origin/HEAD, no local main or master)")
}

// BranchExists checks whether a local branch with the given name exists.
func BranchExists(name string) bool {
	_, err := Run("rev-parse", "--verify", "--quiet", "refs/heads/"+name)
	return err == nil
}

// RemoteBranchExists checks whether origin/<name> exists.
func RemoteBranchExists(name string) bool {
	_, err := Run("rev-parse", "--verify", "--quiet", "refs/remotes/origin/"+name)
	return err == nil
}

// HasUpstream reports whether the given branch has an upstream tracking branch.
func HasUpstream(branch string) bool {
	_, err := Run("rev-parse", "--abbrev-ref", branch+"@{upstream}")
	return err == nil
}

// FetchAll fetches from origin and prunes deleted remote branches.
func FetchAll() error {
	_, err := Run("fetch", "origin", "--prune")
	return err
}

// PullBranch fast-forwards the given branch from origin, whether or not
// it's currently checked out.
//
// `git pull` always merges into whatever is *currently checked out*, so
// calling it with a branch name that isn't checked out does not do what
// it looks like it does - it can even merge unrelated history into the
// wrong branch. To update a branch that isn't checked out, we instead
// fast-forward its local ref directly via a fetch refspec, which is
// itself fast-forward-only unless force-prefixed with "+".
func PullBranch(branch string) error {
	current, err := CurrentBranch()
	if err == nil && current == branch {
		if !HasUpstream(branch) {
			return nil
		}
		_, err := Run("pull", "--ff-only", "origin", branch)
		return err
	}
	if !RemoteBranchExists(branch) {
		return nil // nothing on origin to update from
	}
	_, err = Run("fetch", "origin", branch+":"+branch)
	return err
}

// Push pushes the current HEAD, setting upstream automatically if missing.
func Push(branch string) error {
	if HasUpstream(branch) {
		_, err := Run("push", "origin", branch)
		return err
	}
	_, err := Run("push", "-u", "origin", branch)
	return err
}

// ForcePush force-pushes (with lease, safer than plain --force) the given branch.
func ForcePush(branch string) error {
	_, err := Run("push", "--force-with-lease", "origin", branch)
	return err
}

// CheckoutBranch switches the working tree to the given branch.
func CheckoutBranch(name string) error {
	_, err := Run("checkout", name)
	return err
}

// CreateBranch creates `name` from `base` (creating base locally from
// origin first if needed), checks it out, and pushes it to origin so
// local and remote are in sync immediately.
func CreateBranch(name, base string) error {
	if err := FetchAll(); err != nil {
		return err
	}
	if !BranchExists(base) {
		if RemoteBranchExists(base) {
			if _, err := Run("checkout", "-b", base, "origin/"+base); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("base branch %q not found locally or on origin", base)
		}
	} else {
		if _, err := Run("checkout", base); err != nil {
			return err
		}
		_ = PullBranch(base)
	}
	if _, err := Run("checkout", "-b", name); err != nil {
		return err
	}
	_, err := Run("push", "-u", "origin", name)
	return err
}

// DeleteLocalBranch deletes a local branch (force, since it's typically
// already merged by the time giteasy offers to delete it).
func DeleteLocalBranch(name string) error {
	_, err := Run("branch", "-D", name)
	return err
}

// DeleteRemoteBranch deletes origin/<name>.
func DeleteRemoteBranch(name string) error {
	_, err := Run("push", "origin", "--delete", name)
	return err
}

// ListLocalBranches returns all local branch names.
func ListLocalBranches() ([]string, error) {
	out, err := Run("for-each-ref", "--format=%(refname:short)", "refs/heads/")
	if err != nil {
		return nil, err
	}
	return splitLines(out), nil
}

// ListAllBranchNames returns the union of local branches and remote
// branches on origin (without the "origin/" prefix, deduplicated,
// excluding HEAD).
func ListAllBranchNames() ([]string, error) {
	local, err := Run("for-each-ref", "--format=%(refname:short)", "refs/heads/")
	if err != nil {
		return nil, err
	}
	remote, err := Run("for-each-ref", "--format=%(refname:short)", "refs/remotes/origin/")
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var result []string
	for _, b := range splitLines(local) {
		if !seen[b] {
			seen[b] = true
			result = append(result, b)
		}
	}
	for _, b := range splitLines(remote) {
		b = strings.TrimPrefix(b, "origin/")
		if b == "HEAD" || seen[b] {
			continue
		}
		seen[b] = true
		result = append(result, b)
	}
	sort.Strings(result)
	return result, nil
}

// ListTags returns all tags, sorted by creation order (oldest first)
// as git already tracks them.
func ListTags() ([]string, error) {
	out, err := Run("tag", "--sort=-v:refname")
	if err != nil {
		return nil, err
	}
	return splitLines(out), nil
}

var semverRe = regexp.MustCompile(`^([vV]?)(\d+)\.(\d+)\.(\d+)(.*)$`)

// LatestTag returns the highest semver-looking tag, or "" if none exist.
func LatestTag() (string, error) {
	tags, err := ListTags()
	if err != nil {
		return "", err
	}
	best := ""
	var bestMaj, bestMin, bestPatch int
	for _, t := range tags {
		m := semverRe.FindStringSubmatch(t)
		if m == nil {
			continue
		}
		maj, _ := strconv.Atoi(m[2])
		min, _ := strconv.Atoi(m[3])
		patch, _ := strconv.Atoi(m[4])
		if best == "" || maj > bestMaj || (maj == bestMaj && min > bestMin) ||
			(maj == bestMaj && min == bestMin && patch > bestPatch) {
			best, bestMaj, bestMin, bestPatch = t, maj, min, patch
		}
	}
	return best, nil
}

// SuggestNextPatch bumps the patch component of a tag like "v1.0.1" -> "v1.0.2".
// If tag doesn't parse as semver, it returns "v1.0.0" as a sane starting point.
func SuggestNextPatch(tag string) string {
	if tag == "" {
		return "v1.0.0"
	}
	m := semverRe.FindStringSubmatch(tag)
	if m == nil {
		return "v1.0.0"
	}
	prefix := m[1]
	maj, _ := strconv.Atoi(m[2])
	min, _ := strconv.Atoi(m[3])
	patch, _ := strconv.Atoi(m[4])
	return fmt.Sprintf("%s%d.%d.%d", prefix, maj, min, patch+1)
}

// CreateTag creates an annotated tag at HEAD.
func CreateTag(name string) error {
	_, err := Run("tag", "-a", name, "-m", name)
	return err
}

// PushTag pushes a single tag to origin.
func PushTag(name string) error {
	_, err := Run("push", "origin", name)
	return err
}

// RebaseOnto rebases the current branch onto the given branch.
// Returns a boolean indicating whether a conflict occurred (in which
// case the caller should surface --continue/--abort instructions
// rather than treating it as a hard failure).
func RebaseOnto(target string) (conflict bool, err error) {
	out, errOut, runErr := runCapture("rebase", target)
	if runErr != nil {
		combined := strings.ToLower(out + errOut)
		if strings.Contains(combined, "conflict") {
			return true, fmt.Errorf("rebase conflict")
		}
		msg := strings.TrimSpace(errOut)
		if msg == "" {
			msg = strings.TrimSpace(out)
		}
		return false, fmt.Errorf("%s", msg)
	}
	return false, nil
}

// MergeBranch merges `source` into the currently checked out branch (target).
func MergeBranch(source string) (conflict bool, err error) {
	out, errOut, runErr := runCapture("merge", "--no-ff", source)
	if runErr != nil {
		combined := strings.ToLower(out + errOut)
		if strings.Contains(combined, "conflict") {
			return true, fmt.Errorf("merge conflict")
		}
		msg := strings.TrimSpace(errOut)
		if msg == "" {
			msg = strings.TrimSpace(out)
		}
		return false, fmt.Errorf("%s", msg)
	}
	return false, nil
}

// ResetMode selects how ResetTo treats the working tree / index.
type ResetMode string

const (
	ResetSoft  ResetMode = "--soft"
	ResetMixed ResetMode = "--mixed"
	ResetHard  ResetMode = "--hard"
)

// ResetTo runs `git reset <mode> <target>`.
func ResetTo(mode ResetMode, target string) error {
	_, err := Run("reset", string(mode), target)
	return err
}

// CommitsAhead returns how many commits the current branch has that
// `base` does not (i.e. commits since diverging from base).
func CommitsAhead(base string) (int, error) {
	out, err := Run("rev-list", "--count", base+"..HEAD")
	if err != nil {
		return 0, err
	}
	n, err := strconv.Atoi(out)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// SquashLastN soft-resets the last n commits and creates one new
// commit with the given message.
func SquashLastN(n int, message string) error {
	if n < 2 {
		return fmt.Errorf("need at least 2 commits to squash")
	}
	target := fmt.Sprintf("HEAD~%d", n)
	if err := ResetTo(ResetSoft, target); err != nil {
		return err
	}
	_, err := Run("commit", "-m", message)
	return err
}

// ConfigUserName returns the configured git user.name, or "" if unset.
func ConfigUserName() string {
	out, _ := Run("config", "user.name")
	return out
}

// IsWorkingTreeClean reports whether there are no staged/unstaged changes.
func IsWorkingTreeClean() bool {
	out, err := Run("status", "--porcelain")
	return err == nil && out == ""
}

func splitLines(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	lines := strings.Split(s, "\n")
	var out []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" {
			out = append(out, l)
		}
	}
	return out
}
