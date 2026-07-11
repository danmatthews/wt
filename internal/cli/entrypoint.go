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

func newEntryPoint() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "entry-point",
		Aliases: []string{"ep"},
		Short:   "Manage the current worktree's entry points",
	}
	cmd.AddCommand(newEntryPointAdd(), newEntryPointSet(), newEntryPointRemove())
	return cmd
}

// updateEntryPoint resolves the current worktree, opens its project under lock,
// and hands the worktree to fn. It centralizes the not-registered check and the
// worktree UpdatedAt bump.
func updateEntryPoint(fn func(w *model.Worktree) error) (*model.Worktree, error) {
	loc, err := gitutil.Resolve()
	if err != nil {
		return nil, err
	}
	st, err := store.Default()
	if err != nil {
		return nil, err
	}
	var w *model.Worktree
	_, err = st.Update(loc.Main, func(p *model.Project) error {
		w = p.FindByPath(loc.Worktree)
		if w == nil {
			return apperr.New(apperr.CodeWorktreeNotRegistered,
				"this worktree is not registered; run `wt register` first")
		}
		if err := fn(w); err != nil {
			return err
		}
		w.UpdatedAt = model.Now()
		return nil
	})
	return w, err
}

func newEntryPointAdd() *cobra.Command {
	var epType, url, description string
	cmd := &cobra.Command{
		Use:   "add <name> --type=url --url=<url> [--description <desc>]",
		Short: "Attach an entry point to the current worktree",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			name := args[0]
			if epType != model.TypeURL {
				return apperr.New(apperr.CodeUnknownEntryPointType,
					"unsupported entry-point type %q (only %q is supported)", epType, model.TypeURL).
					WithDetail("type", epType)
			}
			if url == "" {
				return apperr.New(apperr.CodeUsage, "--url is required for type url")
			}
			w, err := updateEntryPoint(func(w *model.Worktree) error {
				if w.FindEntryPoint(name) != nil {
					return apperr.New(apperr.CodeEntryPointNameConflict,
						"entry point %q already exists on this worktree", name).
						WithDetail("name", name)
				}
				now := model.Now()
				w.EntryPoints = append(w.EntryPoints, &model.EntryPoint{
					Name: name, Type: epType, URL: url, Description: description,
					AddedAt: now, UpdatedAt: now,
				})
				return nil
			})
			if err != nil {
				return err
			}
			output.Emit(w.FindEntryPoint(name), nil, func() {
				fmt.Printf("added entry point %q (%s) → %s\n", name, epType, url)
			})
			return nil
		},
	}
	cmd.Flags().StringVar(&epType, "type", model.TypeURL, "entry-point type")
	cmd.Flags().StringVar(&url, "url", "", "URL for a url entry point (e.g. mysite.test)")
	cmd.Flags().StringVar(&description, "description", "", "what this entry point is for")
	return cmd
}

func newEntryPointSet() *cobra.Command {
	var newName, url, description string
	cmd := &cobra.Command{
		Use:   "set <name> [--name <new>] [--url <url>] [--description <desc>]",
		Short: "Update an entry point on the current worktree",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			var updated *model.EntryPoint
			w, err := updateEntryPoint(func(w *model.Worktree) error {
				ep := w.FindEntryPoint(name)
				if ep == nil {
					return apperr.New(apperr.CodeEntryPointNotFound,
						"no entry point named %q on this worktree", name).
						WithDetail("name", name)
				}
				if cmd.Flags().Changed("name") {
					if newName != name && w.FindEntryPoint(newName) != nil {
						return apperr.New(apperr.CodeEntryPointNameConflict,
							"entry point %q already exists on this worktree", newName).
							WithDetail("name", newName)
					}
					ep.Name = newName
				}
				if cmd.Flags().Changed("url") {
					ep.URL = url
				}
				if cmd.Flags().Changed("description") {
					ep.Description = description
				}
				ep.UpdatedAt = model.Now()
				updated = ep
				return nil
			})
			if err != nil {
				return err
			}
			output.Emit(updated, nil, func() {
				fmt.Printf("updated entry point %q on %q\n", updated.Name, w.Name)
			})
			return nil
		},
	}
	cmd.Flags().StringVar(&newName, "name", "", "rename the entry point")
	cmd.Flags().StringVar(&url, "url", "", "new URL")
	cmd.Flags().StringVar(&description, "description", "", "new description")
	return cmd
}

func newEntryPointRemove() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "Detach an entry point from the current worktree",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			name := args[0]
			w, err := updateEntryPoint(func(w *model.Worktree) error {
				if !w.RemoveEntryPoint(name) {
					return apperr.New(apperr.CodeEntryPointNotFound,
						"no entry point named %q on this worktree", name).
						WithDetail("name", name)
				}
				return nil
			})
			if err != nil {
				return err
			}
			output.Emit(map[string]any{"name": name, "worktree": w.Name}, nil, func() {
				fmt.Printf("removed entry point %q from %q\n", name, w.Name)
			})
			return nil
		},
	}
}
