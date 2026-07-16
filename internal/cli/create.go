package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/danmatthews/wt/internal/apperr"
	"github.com/danmatthews/wt/internal/gitutil"
	"github.com/danmatthews/wt/internal/output"
	"github.com/danmatthews/wt/internal/store"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

func newCreate() *cobra.Command {
	var name, path, branch, description, app string
	cmd := &cobra.Command{
		Use:   "create [--name <name>] [--path <path>] [--branch <branch>] [--description <desc>] [--app <app>]",
		Short: "Create a git worktree and register it in one step",
		Long: "Runs `git worktree add` on a new branch and records the result in " +
			"the registry. Unlike `register`, which adopts a worktree that already " +
			"exists, `create` is the one command that mutates git (ADR 0002). Run " +
			"it from inside the project. With no --name it prompts for one.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runCreate(name, path, branch, description, app,
				cmd.Flags().Changed("description"), cmd.Flags().Changed("app"))
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "friendly name for the worktree (prompted if omitted)")
	cmd.Flags().StringVar(&path, "path", "", "where to create the worktree (default: sibling of the main worktree named after --name)")
	cmd.Flags().StringVar(&branch, "branch", "", "name of the new branch to check out (default: --name)")
	cmd.Flags().StringVar(&description, "description", "", "what this worktree is for")
	cmd.Flags().StringVar(&app, "app", "", "application used to work on this worktree, e.g. \"Conductor.app\"")
	return cmd
}

func runCreate(name, path, branch, description, app string, descSet, appSet bool) error {
	if err := gitutil.Available(); err != nil {
		return err
	}
	// Resolve the enclosing project: we need the main worktree to derive the
	// default path/branch and to key the registry entry.
	loc, err := gitutil.Resolve()
	if err != nil {
		return err
	}

	if name == "" {
		if name, err = promptName(); err != nil {
			return err
		}
	}

	// Fail before touching git if the name is already taken in this project, so
	// we never leave an orphaned worktree we can't register (ADR 0004).
	st, err := store.Default()
	if err != nil {
		return err
	}
	if p, err := st.Project(loc.Main); err != nil {
		return err
	} else if p != nil {
		if ex := p.FindByName(name); ex != nil {
			return apperr.New(apperr.CodeNameConflict,
				"name %q is already used by %s in this project", name, ex.Path).
				WithDetail("name", name).WithDetail("existing_path", ex.Path)
		}
	}

	if branch == "" {
		branch = name
	}
	if path == "" {
		path = filepath.Join(filepath.Dir(loc.Main), name)
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return apperr.New(apperr.CodeIOError, "cannot resolve path %q: %s", path, err)
	}

	created, err := gitutil.AddWorktree(abs, branch)
	if err != nil {
		return err
	}

	newLoc := gitutil.Location{Worktree: created, Main: loc.Main}
	result, err := registerWorktree(newLoc, name, description, app, descSet, appSet)
	if err != nil {
		// The worktree exists on disk; make that explicit so the user can adopt
		// it with `wt register` rather than assume nothing happened.
		return apperr.New(apperr.CodeIOError,
			"worktree created at %s but registration failed: %s", created, err.Error()).
			WithDetail("path", created)
	}

	output.Emit(result, map[string]any{"branch": branch}, func() {
		fmt.Printf("created worktree %q on branch %q at %s\n", result.Name, branch, result.Path)
	})
	return nil
}

// promptName asks for a worktree name on the terminal. It refuses when it
// cannot prompt (--json or a non-interactive stdin), pointing the caller at
// --name instead.
func promptName() (string, error) {
	if output.JSON {
		return "", apperr.New(apperr.CodeUsage, "--name is required with --json")
	}
	if !isatty.IsTerminal(os.Stdin.Fd()) {
		return "", apperr.New(apperr.CodeUsage, "--name is required when stdin is not a terminal")
	}
	fmt.Fprint(os.Stderr, "worktree name: ")
	line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	name := strings.TrimSpace(line)
	if name == "" {
		return "", apperr.New(apperr.CodeUsage, "no name provided")
	}
	return name, nil
}
