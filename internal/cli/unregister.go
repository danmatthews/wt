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

func newUnregister() *cobra.Command {
	return &cobra.Command{
		Use:   "unregister",
		Short: "Remove the current worktree's registry entry (does not touch the worktree on disk)",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runUnregister()
		},
	}
}

func runUnregister() error {
	loc, err := gitutil.Resolve()
	if err != nil {
		return err
	}
	st, err := store.Default()
	if err != nil {
		return err
	}

	if _, err := st.Update(loc.Main, func(p *model.Project) error {
		if !p.RemoveByPath(loc.Worktree) {
			return apperr.New(apperr.CodeWorktreeNotRegistered,
				"this worktree is not registered")
		}
		return nil
	}); err != nil {
		return err
	}

	output.Emit(map[string]any{"path": loc.Worktree}, nil, func() {
		fmt.Printf("unregistered %s\n", loc.Worktree)
	})
	return nil
}
