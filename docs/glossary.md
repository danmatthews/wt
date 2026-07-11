# Glossary — `wt`

The ubiquitous language for `wt`. See `docs/DESIGN.md` for the consolidated design and `docs/adr/` for the decision record.

| Term | Definition |
|------|------------|
| **wt** | A local, single-user registry of git worktrees, queryable by agents/tools to see what work is in flight per project. Records worktrees; never creates/deletes them. |
| **Registry** | The datastore: one TOML file per project under `~/.config/wt/projects/`, local to one macOS user account, guarded by per-file locking. |
| **Project** | The base git repository worktrees belong to. Identity = the absolute path of its **main worktree** on the local machine. |
| **Main worktree** | The primary checkout holding the shared `.git`; its path identifies the project. Registrable as a worktree entry, but flagged **special**. |
| **Worktree entry** | A registered record for one git worktree. Identity = the worktree's absolute path. Holds `name`, optional `description`, a `special` flag, and zero or more entry points. |
| **Name** | Mutable label, unique **within a project**; the CLI/`list` handle for a worktree. Not globally unique, not the identity. |
| **Special (base)** | Flag on the worktree entry that is the project's main worktree, so consumers can distinguish the base checkout from linked worktrees. |
| **Entry point** | An explicit, typed, individually-named access point on a worktree. Identity = its name within that worktree. May carry its own description. A worktree may have many, including several of the same type. |
| **URL entry point** | An entry point of type `url` holding a caller-provided address (e.g. `mysite.test`). `wt` does not derive it and knows nothing of Herd. |
| **Active** | A worktree whose entry exists and whose path still exists on disk. |
| **Stale entry** | A worktree entry whose path no longer exists on disk. Removed via `unregister` or auto-prune (on `list`). |
| **cwd self-identification** | `wt` resolves the enclosing worktree and its project from the current directory via git plumbing; mutating commands require being inside a worktree. |

## Command summary

| Command | Scope | Purpose |
|---------|-------|---------|
| `wt register --name <n> [--description <d>]` | cwd worktree | Create/update this worktree (idempotent by path). |
| `wt set name \| description "<v>"` | cwd worktree | Rename / re-describe this worktree. |
| `wt unregister` | cwd worktree | Remove this worktree's entry. |
| `wt entry-point add \| set \| remove <ep> …` | cwd worktree | Manage this worktree's typed entry points. |
| `wt list [--wt]` | global / cwd | List all projects→worktrees→entry points, or just the current worktree. |

All commands support `--json`. Mutating commands fail outside a worktree.
