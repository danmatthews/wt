package store

import (
	"fmt"
	"sync"
	"testing"

	"github.com/danmatthews/wt/internal/model"
)

// TestConcurrentUpdatesNoLostWrites is the core correctness claim behind ADR
// 0010: concurrent writers to the same project must serialize so no update is
// lost. N goroutines each append a distinct worktree; all N must survive.
func TestConcurrentUpdatesNoLostWrites(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	st, err := Default()
	if err != nil {
		t.Fatalf("Default: %v", err)
	}

	const project = "/tmp/proj"
	const n = 25

	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := st.Update(project, func(p *model.Project) error {
				p.Worktrees = append(p.Worktrees, &model.Worktree{
					Path: fmt.Sprintf("/tmp/proj-wt-%d", i),
					Name: fmt.Sprintf("wt-%d", i),
				})
				return nil
			})
			if err != nil {
				t.Errorf("Update %d: %v", i, err)
			}
		}(i)
	}
	wg.Wait()

	p, _, err := loadProject(st, project)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if got := len(p.Worktrees); got != n {
		t.Fatalf("lost updates: got %d worktrees, want %d", got, n)
	}
}

// TestUpdateDeletesEmptyProject verifies a project file is removed once its
// last worktree is gone.
func TestUpdateDeletesEmptyProject(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	st, err := Default()
	if err != nil {
		t.Fatalf("Default: %v", err)
	}
	const project = "/tmp/proj"

	if _, err := st.Update(project, func(p *model.Project) error {
		p.Worktrees = append(p.Worktrees, &model.Worktree{Path: "/tmp/w", Name: "w"})
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := st.Update(project, func(p *model.Project) error {
		p.RemoveByPath("/tmp/w")
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	all, err := st.All()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 0 {
		t.Fatalf("expected empty registry, got %d projects", len(all))
	}
}

// loadProject is a test helper that reads a project back through the same code
// path Update uses.
func loadProject(s *Store, projectPath string) (*model.Project, bool, error) {
	p, err := readFile(s.fileFor(projectPath))
	return p, p != nil, err
}
