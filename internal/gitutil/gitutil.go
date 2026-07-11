// Package gitutil resolves worktree identities from the current directory
// via git plumbing (ADR 0007). wt never mutates git (ADR 0002).
package gitutil

import (
	"bufio"
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
