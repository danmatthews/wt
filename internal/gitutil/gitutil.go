// Package gitutil resolves worktree identities from the current directory
// via git plumbing (ADR 0007). It is read-only except for AddWorktree, which
// creates a worktree for `wt create` — the sanctioned exception to ADR 0002
// (the registry otherwise never mutates git).
package gitutil

import (
	"bufio"
	"bytes"
	"os/exec"
	"strings"

	"github.com/danmatthews/wt/internal/apperr"
)

// Location describes where wt is being invoked from.
type Location struct {
	Worktree string // absolute path of the enclosing worktree
	Main     string // absolute path of the repo's main worktree (project identity)
}

// IsMain reports whether the current worktree is the repo's main worktree.
func (l Location) IsMain() bool { return l.Worktree == l.Main }

func git(args ...string) (string, error) {
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// run executes git in dir (or cwd when dir is empty), returning trimmed stdout.
// On failure it surfaces git's stderr so callers can report why it failed.
func run(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		if msg := strings.TrimSpace(stderr.String()); msg != "" {
			return "", apperr.New(apperr.CodeGitError, "%s", msg)
		}
		return "", apperr.New(apperr.CodeGitError, "git %s: %s", strings.Join(args, " "), err)
	}
	return strings.TrimSpace(stdout.String()), nil
}

// Available reports whether the git executable is on PATH, returning a
// git_unavailable error otherwise.
func Available() error {
	if _, err := exec.LookPath("git"); err != nil {
		return apperr.New(apperr.CodeGitUnavailable,
			"git executable not found on PATH")
	}
	return nil
}

// RemoveWorktree deletes the worktree at path via `git worktree remove`. With
// force, git also removes a worktree that has modified or untracked files (and
// unlocks a locked one). This and AddWorktree are the only places wt mutates
// git (ADR 0002).
func RemoveWorktree(path string, force bool) error {
	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, path)
	_, err := run("", args...)
	return err
}

// AddWorktree creates a worktree at path checked out on a new branch, and
// returns the canonical toplevel path of the created worktree (matching what
// Resolve would report from inside it, so registry identity stays consistent —
// ADR 0003). This is the one place wt mutates git (ADR 0002).
func AddWorktree(path, branch string) (string, error) {
	if _, err := run("", "worktree", "add", path, "-b", branch); err != nil {
		return "", err
	}
	top, err := run(path, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", apperr.New(apperr.CodeIOError,
			"worktree created at %s but its path could not be resolved: %s", path, err)
	}
	return top, nil
}

// Resolve determines the current worktree and its main worktree. It returns a
// not_in_worktree error when cwd is not inside a git worktree.
func Resolve() (Location, error) {
	top, err := git("rev-parse", "--show-toplevel")
	if err != nil || top == "" {
		return Location{}, apperr.New(apperr.CodeNotInWorktree,
			"not inside a git worktree")
	}
	main, err := mainWorktree()
	if err != nil {
		return Location{}, apperr.New(apperr.CodeIOError,
			"could not resolve main worktree: %s", err.Error())
	}
	return Location{Worktree: top, Main: main}, nil
}

// mainWorktree returns the first worktree listed by `git worktree list`, which
// is always the repo's main worktree.
func mainWorktree() (string, error) {
	out, err := git("worktree", "list", "--porcelain")
	if err != nil {
		return "", err
	}
	sc := bufio.NewScanner(strings.NewReader(out))
	for sc.Scan() {
		line := sc.Text()
		if path, ok := strings.CutPrefix(line, "worktree "); ok {
			return path, nil
		}
	}
	return "", apperr.New(apperr.CodeIOError, "no worktree found in `git worktree list`")
}
