// Package model defines the persisted domain types for the wt registry.
//
// One Project is stored per TOML file (ADR 0010). Identities are absolute
// paths (ADR 0003): a Project is keyed by its main-worktree path and a
// Worktree by its own path. Names are mutable labels (ADR 0004, 0008).
package model

import "time"

// Project is the base git repository. It is the unit of storage.
type Project struct {
	ProjectPath string      `toml:"project_path" json:"project_path"`
	Worktrees   []*Worktree `toml:"worktree" json:"worktrees"`
}

// Worktree is one registered git worktree belonging to a Project.
type Worktree struct {
	Path         string        `toml:"path" json:"path"`
	Name         string        `toml:"name" json:"name"`
	Description  string        `toml:"description,omitempty" json:"description,omitempty"`
	Special      bool          `toml:"special,omitempty" json:"special"`
	RegisteredAt string        `toml:"registered_at" json:"registered_at"`
	UpdatedAt    string        `toml:"updated_at" json:"updated_at"`
	EntryPoints  []*EntryPoint `toml:"entry_point,omitempty" json:"entry_points"`
}

// EntryPoint is a typed, individually-named access point on a Worktree
// (ADR 0005, 0008). The first type is "url".
type EntryPoint struct {
	Name        string `toml:"name" json:"name"`
	Type        string `toml:"type" json:"type"`
	URL         string `toml:"url,omitempty" json:"url,omitempty"`
	Description string `toml:"description,omitempty" json:"description,omitempty"`
	AddedAt     string `toml:"added_at" json:"added_at"`
	UpdatedAt   string `toml:"updated_at" json:"updated_at"`
}

// TypeURL is the only entry-point type supported today.
const TypeURL = "url"

// Now returns the current instant as an ISO-8601 UTC string (ADR 0014).
func Now() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// FindByPath returns the worktree with the given absolute path, or nil.
func (p *Project) FindByPath(path string) *Worktree {
	for _, w := range p.Worktrees {
		if w.Path == path {
			return w
		}
	}
	return nil
}

// FindByName returns the worktree with the given name, or nil. Names are
// unique within a project (ADR 0004).
func (p *Project) FindByName(name string) *Worktree {
	for _, w := range p.Worktrees {
		if w.Name == name {
			return w
		}
	}
	return nil
}

// RemoveByPath drops the worktree with the given path, reporting whether one
// was removed.
func (p *Project) RemoveByPath(path string) bool {
	for i, w := range p.Worktrees {
		if w.Path == path {
			p.Worktrees = append(p.Worktrees[:i], p.Worktrees[i+1:]...)
			return true
		}
	}
	return false
}

// FindEntryPoint returns the entry point with the given name, or nil.
func (w *Worktree) FindEntryPoint(name string) *EntryPoint {
	for _, ep := range w.EntryPoints {
		if ep.Name == name {
			return ep
		}
	}
	return nil
}

// RemoveEntryPoint drops the entry point with the given name, reporting
// whether one was removed.
func (w *Worktree) RemoveEntryPoint(name string) bool {
	for i, ep := range w.EntryPoints {
		if ep.Name == name {
			w.EntryPoints = append(w.EntryPoints[:i], w.EntryPoints[i+1:]...)
			return true
		}
	}
	return false
}
