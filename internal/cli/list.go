package cli

import (
	"fmt"
	"os"

	"github.com/danmatthews/wt/internal/apperr"
	"github.com/danmatthews/wt/internal/gitutil"
	"github.com/danmatthews/wt/internal/model"
	"github.com/danmatthews/wt/internal/output"
	"github.com/danmatthews/wt/internal/store"
	"github.com/spf13/cobra"
)

// pruned records a stale worktree entry removed during a list (ADR 0006).
type pruned struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	ProjectPath string `json:"project_path"`
	Reason      string `json:"reason"`
}

func newList() *cobra.Command {
	var wtOnly bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List registered projects, worktrees and entry points",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runList(wtOnly)
		},
	}
	cmd.Flags().BoolVar(&wtOnly, "wt", false, "show only the current worktree")
	return cmd
}

func runList(wtOnly bool) error {
	st, err := store.Default()
	if err != nil {
		return err
	}
	if wtOnly {
		return listCurrent(st)
	}
	return listAll(st)
}

func listAll(st *store.Store) error {
	projects, err := st.All()
	if err != nil {
		return err
	}
	var allPruned []pruned
	var display []*model.Project
	for _, p := range projects {
		post, err := st.Update(p.ProjectPath, func(pp *model.Project) error {
			allPruned = append(allPruned, pruneStale(pp)...)
			return nil
		})
		if err != nil {
			return err
		}
		if len(post.Worktrees) > 0 {
			display = append(display, post)
		}
	}

	output.Emit(display, prunedExtra(allPruned), func() {
		reportPruned(allPruned)
		if len(display) == 0 {
			fmt.Println(styleMuted.Render("no worktrees registered"))
			return
		}
		for i, p := range display {
			if i > 0 {
				fmt.Println()
			}
			renderProject(p)
		}
	})
	return nil
}

func listCurrent(st *store.Store) error {
	loc, err := gitutil.Resolve()
	if err != nil {
		return err
	}
	var current *model.Worktree
	var pr []pruned
	if _, err := st.Update(loc.Main, func(pp *model.Project) error {
		pr = pruneStale(pp)
		current = pp.FindByPath(loc.Worktree)
		return nil
	}); err != nil {
		return err
	}
	if current == nil {
		return apperr.New(apperr.CodeWorktreeNotRegistered,
			"this worktree is not registered")
	}

	output.Emit(current, prunedExtra(pr), func() {
		reportPruned(pr)
		renderWorktree(current)
	})
	return nil
}

// pruneStale removes worktrees whose path no longer exists on disk and returns
// the removed entries (ADR 0006).
func pruneStale(p *model.Project) []pruned {
	var removed []pruned
	kept := p.Worktrees[:0]
	for _, w := range p.Worktrees {
		if _, err := os.Stat(w.Path); os.IsNotExist(err) {
			removed = append(removed, pruned{
				Name: w.Name, Path: w.Path, ProjectPath: p.ProjectPath, Reason: "path_gone",
			})
			continue
		}
		kept = append(kept, w)
	}
	p.Worktrees = kept
	return removed
}

func prunedExtra(pr []pruned) map[string]any {
	if pr == nil {
		pr = []pruned{}
	}
	return map[string]any{"pruned": pr}
}

func reportPruned(pr []pruned) {
	for _, p := range pr {
		fmt.Fprintf(os.Stderr, "pruned %q (%s)\n", p.Name, p.Reason)
	}
}
