// Package store persists projects as one TOML file each under ~/.config/wt
// (ADR 0010). Concurrent writers are serialized per-project with an advisory
// flock, and writes are atomic (temp file + rename) so a file is never seen
// half-written.
package store

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/danmatthews/wt/internal/apperr"
	"github.com/danmatthews/wt/internal/model"
)

// lockTimeout bounds how long we wait for a contended project lock before
// returning lock_timeout (ADR 0015).
const lockTimeout = 5 * time.Second

// Store is the on-disk registry rooted at ~/.config/wt.
type Store struct {
	root string
}

// Default returns the registry under $XDG_CONFIG_HOME/wt (or ~/.config/wt),
// creating the projects directory if needed.
func Default() (*Store, error) {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, apperr.New(apperr.CodeIOError, "cannot find home directory: %s", err)
		}
		base = filepath.Join(home, ".config")
	}
	s := &Store{root: filepath.Join(base, "wt")}
	if err := os.MkdirAll(s.projectsDir(), 0o755); err != nil {
		return nil, apperr.New(apperr.CodeIOError, "cannot create registry dir: %s", err)
	}
	return s, nil
}

func (s *Store) projectsDir() string { return filepath.Join(s.root, "projects") }

// fileFor maps a project's main-worktree path to its TOML file. The filename
// is a hash of the path (ADR 0010); the human path lives inside the file.
func (s *Store) fileFor(projectPath string) string {
	sum := sha256.Sum256([]byte(projectPath))
	return filepath.Join(s.projectsDir(), hex.EncodeToString(sum[:8])+".toml")
}

// Update runs fn under the project's exclusive lock, passing the loaded
// project (or a fresh one keyed to projectPath). After fn returns, the project
// is written atomically — or its file deleted if it has no worktrees left.
func (s *Store) Update(projectPath string, fn func(*model.Project) error) (*model.Project, error) {
	file := s.fileFor(projectPath)
	unlock, err := lock(file)
	if err != nil {
		return nil, err
	}
	defer unlock()

	p, err := readFile(file)
	if err != nil {
		return nil, err
	}
	if p == nil {
		p = &model.Project{ProjectPath: projectPath}
	}
	p.ProjectPath = projectPath

	if err := fn(p); err != nil {
		return nil, err
	}

	if len(p.Worktrees) == 0 {
		if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
			return nil, apperr.New(apperr.CodeIOError, "cannot remove empty project file: %s", err)
		}
		return p, nil
	}
	if err := writeAtomic(file, p); err != nil {
		return nil, err
	}
	return p, nil
}

// All returns every stored project (unlocked snapshot read).
func (s *Store) All() ([]*model.Project, error) {
	entries, err := os.ReadDir(s.projectsDir())
	if err != nil {
		return nil, apperr.New(apperr.CodeIOError, "cannot read registry: %s", err)
	}
	var projects []*model.Project
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".toml" {
			continue
		}
		p, err := readFile(filepath.Join(s.projectsDir(), e.Name()))
		if err != nil {
			return nil, err
		}
		if p != nil {
			projects = append(projects, p)
		}
	}
	return projects, nil
}

// readFile loads a project TOML file, returning (nil, nil) if it is absent.
func readFile(file string) (*model.Project, error) {
	data, err := os.ReadFile(file)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, apperr.New(apperr.CodeIOError, "cannot read project file: %s", err)
	}
	var p model.Project
	if err := toml.Unmarshal(data, &p); err != nil {
		return nil, apperr.New(apperr.CodeIOError, "corrupt project file %s: %s", file, err)
	}
	return &p, nil
}

// writeAtomic serializes p to a temp file in the same directory, then renames
// it over the target so readers never observe a partial write.
func writeAtomic(file string, p *model.Project) error {
	tmp, err := os.CreateTemp(filepath.Dir(file), ".wt-*.tmp")
	if err != nil {
		return apperr.New(apperr.CodeIOError, "cannot create temp file: %s", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // no-op after a successful rename

	enc := toml.NewEncoder(tmp)
	if err := enc.Encode(p); err != nil {
		tmp.Close()
		return apperr.New(apperr.CodeIOError, "cannot encode project: %s", err)
	}
	if err := tmp.Close(); err != nil {
		return apperr.New(apperr.CodeIOError, "cannot flush project file: %s", err)
	}
	if err := os.Rename(tmpName, file); err != nil {
		return apperr.New(apperr.CodeIOError, "cannot commit project file: %s", err)
	}
	return nil
}

// lock takes an exclusive advisory lock on a sidecar <file>.lock, retrying
// until lockTimeout elapses. The returned func releases the lock.
func lock(file string) (func(), error) {
	lf, err := os.OpenFile(file+".lock", os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, apperr.New(apperr.CodeIOError, "cannot open lock file: %s", err)
	}
	deadline := time.Now().Add(lockTimeout)
	for {
		err := syscall.Flock(int(lf.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			return func() {
				syscall.Flock(int(lf.Fd()), syscall.LOCK_UN)
				lf.Close()
			}, nil
		}
		if err != syscall.EWOULDBLOCK {
			lf.Close()
			return nil, apperr.New(apperr.CodeIOError, "cannot lock project: %s", err)
		}
		if time.Now().After(deadline) {
			lf.Close()
			return nil, apperr.New(apperr.CodeLockTimeout, "timed out waiting for project lock")
		}
		time.Sleep(25 * time.Millisecond)
	}
}
