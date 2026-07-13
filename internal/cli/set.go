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

func newSet() *cobra.Command {
	return &cobra.Command{
		Use:       "set <name|description> <value>",
		Short:     "Update this worktree's name or description",
		Args:      cobra.ExactArgs(2),
		ValidArgs: []string{"name", "description"},
		RunE: func(_ *cobra.Command, args []string) error {
			return runSet(args[0], args[1])
		},
	}
}

func runSet(field, value string) error {
	if field != "name" && field != "description" {
		return apperr.New(apperr.CodeUsage, "unknown field %q; expected name or description", field)
	}
	loc, err := gitutil.Resolve()
	if err != nil {
		return err
	}
	st, err := store.Default()
	if err != nil {
		return err
	}

	var result *model.Worktree
	if _, err := st.Update(loc.Main, func(p *model.Project) error {
		w := p.FindByPath(loc.Worktree)
		if w == nil {
			return apperr.New(apperr.CodeWorktreeNotRegistered,
				"this worktree is not registered; run `wt register` first")
		}
		switch field {
		case "name":
			if ex := p.FindByName(value); ex != nil && ex.Path != w.Path {
				return apperr.New(apperr.CodeNameConflict,
					"name %q is already used by %s in this project", value, ex.Path).
					WithDetail("name", value)
			}
			w.Name = value
		case "description":
			w.Description = value
		}
		w.UpdatedAt = model.Now()
		result = w
		return nil
	}); err != nil {
		return err
	}

	output.Emit(result, nil, func() {
		fmt.Printf("set %s for %q\n", field, result.Name)
	})
	return nil
}
