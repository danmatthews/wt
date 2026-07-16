# ADR 0002 — `wt` records worktrees; it never manages git

**Status:** Accepted (amended 2026-07-14 — see Amendment)
**Date:** 2026-07-11

## Context
`wt` could either own worktree creation (`git worktree add/remove`) or purely record worktrees created by other tools. AI tools (Conductor et al.) already have their own worktree-creation workflows, and git worktree syntax has many variants.

## Decision
`wt` is a **registry only**. It records metadata about worktrees that already exist. It never runs `git worktree add`, `git worktree remove`, or any mutating git command.

## Consequences
- `wt register` assumes the worktree already exists on disk.
- No need to model or support the many git worktree invocation styles.
- Removal from the registry (`unregister` / prune) does not delete the actual worktree.
- Creation ownership could be added later without breaking the registry model.

## Amendment (2026-07-14) — `wt create` and `wt remove`
Taking up the "creation ownership could be added later" consequence above, wt now owns the two ends of a worktree's lifecycle as a convenience for the common "spin up a fresh worktree and track it / tear it down" workflow:
- `wt create` runs a single `git worktree add <path> -b <branch>`, then registers the result.
- `wt remove <name>` runs `git worktree remove` on the named worktree, then unregisters it.

These two are the **only** commands that mutate git; `create` and `remove` run exactly one `git worktree` subcommand each and nothing else. The registry model is otherwise unchanged:
- `register` still adopts a worktree that already exists, and `unregister` / prune still only drop the registry entry and never touch disk. `create`/`remove` are the git-touching supersets; the record-only commands remain the default.
- `remove` is destructive, so it is guarded: it confirms interactively (or requires `--force`), refuses the main worktree and the worktree you are standing in, and defers to git's own dirty-tree protection (uncommitted/untracked files need `--force`). It removes from git before unregistering, so a mid-operation failure leaves at worst a stale entry that `wt list` self-heals (ADR 0006) rather than an untracked live worktree.
- We still model only the worktree-add form we drive ourselves (a new branch); arbitrary `git worktree add` variants remain out of scope — use git directly then `wt register`.
