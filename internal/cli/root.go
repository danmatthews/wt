// Package cli wires up the wt command tree.
package cli

import (
	"github.com/danmatthews/wt/internal/output"
	"github.com/spf13/cobra"
)

// Execute runs the root command and returns a process exit code.
func Execute() int {
	root := newRoot()
	if err := root.Execute(); err != nil {
		output.Fail(err)
		return 1
	}
	return 0
}

func newRoot() *cobra.Command {
	root := &cobra.Command{
		Use:           "wt",
		Short:         "A local registry of git worktrees for agents and tools",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.PersistentFlags().BoolVar(&output.JSON, "json", false, "emit machine-readable JSON output")
	root.AddCommand(
		newRegister(),
		newList(),
		newSet(),
		newUnregister(),
		newEntryPoint(),
	)
	return root
}
