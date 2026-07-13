# ADR 0002 — `wt` records worktrees; it never manages git

**Status:** Accepted
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
