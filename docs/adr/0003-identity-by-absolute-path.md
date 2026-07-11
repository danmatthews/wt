# ADR 0003 — Identity is by absolute path; `register` is idempotent

**Status:** Accepted
**Date:** 2026-07-11

## Context
We need stable identities for both "the project" and "a worktree entry" so records can be grouped and re-registration is well-defined.

## Decision
- **Project identity** = the absolute path of the **main worktree** on the local machine.
- **Worktree-entry identity** = the absolute path of the **worktree** on the local machine.
- Because a path *is* a worktree, `wt register` is **idempotent by path**: re-registering the same path updates the existing entry rather than creating a duplicate or erroring.

## Consequences
- `wt` can resolve both identities from any cwd inside a worktree via git plumbing (e.g. the git common dir → main worktree path; the worktree toplevel → worktree path).
- Moving a worktree directory changes its identity (acceptable given the local, single-machine scope).
- Grouping for `wt list` is by project path.
- Enables reliable staleness detection: an entry whose path no longer exists on disk is stale (see ADR 0006).
