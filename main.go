// Command wt is a local, single-user registry of git worktrees for agents and
// tools (see docs/DESIGN.md).
package main

import (
	"os"

	"github.com/danmatthews/wt/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
