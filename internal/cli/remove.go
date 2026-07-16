package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/danmatthews/wt/internal/apperr"
	"github.com/danmatthews/wt/internal/gitutil"
	"github.com/danmatthews/wt/internal/model"
	"github.com/danmatthews/wt/internal/output"
	"github.com/danmatthews/wt/internal/store"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

func newRemove() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "remove <name> [--force]",
		Short: "Delete a worktree and drop its registry entry",
		Long: "Runs `git worktree remove` on the named worktree in the current " +
			"project, then unregisters it. Prompts for confirmation unless " +
			"--force is given. This deletes the worktree from disk — the " +
			"destructive counterpart to `create` (ADR 0002). Target it by name " +
			"and run from another worktree (e.g. the main one); git will not " +
			"remove the worktree you are standing in.",
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runRemove(args[0], force)
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "skip confirmation and remove even if the worktree has uncommitted or untracked changes")
	return cmd
}

func runRemove(name string, force bool) error {
	if err := gitutil.Available(); err != nil {
		return err
	}
	loc, err := gitutil.Resolve()
	if err != nil {
		return err
	}

	st, err := store.Default()
	if err != nil {
		return err
	}
	p, err := st.Project(loc.Main)
	if err != nil {
		return err
	}
	var target *model.Worktree
	if p != nil {
		target = p.FindByName(name)
	}
	if target == nil {
		return apperr.New(apperr.CodeWorktreeNotFound,
			"no worktree named %q in this project", name).WithDetail("name", name)
	}
	if target.Special {
		return apperr.New(apperr.CodeUsage,
			"refusing to remove %q: it is the project's main worktree", name)
	}
	if target.Path == loc.Worktree {
		return apperr.New(apperr.CodeUsage,
			"cannot remove %q: it is the worktree you are in; run from another worktree (e.g. the main worktree)", name)
	}

	if !force {
		if output.JSON {
			return apperr.New(apperr.CodeUsage, "--force is required with --json")
		}
		ok, err := confirm(fmt.Sprintf("remove worktree %q at %s?", name, target.Path))
		if err != nil {
			return err
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "aborted")
			return nil
		}
	}

	// Remove from git first: if unregistering then failed, the entry would be
	// left tracking a live worktree. This ordering instead leaves at worst a
	// stale entry, which `wt list` self-heals (ADR 0006).
	if err := gitutil.RemoveWorktree(target.Path, force); err != nil {
		return err
	}
	if _, err := st.Update(loc.Main, func(pp *model.Project) error {
		pp.RemoveByPath(target.Path)
		return nil
	}); err != nil {
		return err
	}

	output.Emit(map[string]any{"name": name, "path": target.Path}, nil, func() {
		fmt.Printf("removed worktree %q at %s\n", name, target.Path)
	})
	return nil
}

// confirm asks a yes/no question on the terminal, defaulting to no. It refuses
// (usage error) when stdin is not a terminal, pointing the caller at --force.
func confirm(question string) (bool, error) {
	if !isatty.IsTerminal(os.Stdin.Fd()) {
		return false, apperr.New(apperr.CodeUsage,
			"refusing to remove without confirmation on a non-interactive stdin; pass --force")
	}
	fmt.Fprintf(os.Stderr, "%s [y/N] ", question)
	line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	resp := strings.ToLower(strings.TrimSpace(line))
	return resp == "y" || resp == "yes", nil
}
