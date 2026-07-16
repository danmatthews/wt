package cli

import (
	"fmt"

	"github.com/danmatthews/wt/internal/apperr"
	"github.com/danmatthews/wt/internal/gitutil"
	"github.com/danmatthews/wt/internal/model"
	"github.com/danmatthews/wt/internal/output"
	"github.com/danmatthews/wt/internal/store"
	"github.com/spf13/cobra"
)

func newRegister() *cobra.Command {
	var name, description, app string
	cmd := &cobra.Command{
		Use:   "register --name <name> [--description <desc>] [--app <app>]",
		Short: "Register (or update) the current worktree",
		Long: "Records the worktree containing the current directory. Idempotent " +
			"by path: re-running updates the existing entry (ADR 0003).",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runRegister(name, description, app,
				cmd.Flags().Changed("description"), cmd.Flags().Changed("app"))
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "friendly name for this worktree (required)")
	cmd.Flags().StringVar(&description, "description", "", "what this worktree is for")
	cmd.Flags().StringVar(&app, "app", "", "application used to work on this worktree, e.g. \"Conductor.app\"")
	cmd.MarkFlagRequired("name")
	return cmd
}

func runRegister(name, description, app string, descSet, appSet bool) error {
	loc, err := gitutil.Resolve()
	if err != nil {
		return err
	}
	result, err := registerWorktree(loc, name, description, app, descSet, appSet)
	if err != nil {
		return err
	}
	output.Emit(result, nil, func() {
		fmt.Printf("registered %q at %s\n", result.Name, result.Path)
	})
	return nil
}

// registerWorktree records (or updates, idempotent by path — ADR 0003) the
// worktree at loc.Worktree in loc's project. descSet/appSet gate whether the
// respective field is written, so a zero value can clear vs. leave-as-is.
// Shared by `register` (cwd) and `create` (the freshly-created path).
func registerWorktree(loc gitutil.Location, name, description, app string, descSet, appSet bool) (*model.Worktree, error) {
	st, err := store.Default()
	if err != nil {
		return nil, err
	}
	var result *model.Worktree
	if _, err := st.Update(loc.Main, func(p *model.Project) error {
		if ex := p.FindByName(name); ex != nil && ex.Path != loc.Worktree {
			return apperr.New(apperr.CodeNameConflict,
				"name %q is already used by %s in this project", name, ex.Path).
				WithDetail("name", name).WithDetail("existing_path", ex.Path)
		}
		now := model.Now()
		w := p.FindByPath(loc.Worktree)
		if w == nil {
			w = &model.Worktree{Path: loc.Worktree, RegisteredAt: now}
			p.Worktrees = append(p.Worktrees, w)
		}
		w.Name = name
		if descSet {
			w.Description = description
		}
		if appSet {
			w.App = app
		}
		w.Special = loc.IsMain()
		w.UpdatedAt = now
		result = w
		return nil
	}); err != nil {
		return nil, err
	}
	return result, nil
}
