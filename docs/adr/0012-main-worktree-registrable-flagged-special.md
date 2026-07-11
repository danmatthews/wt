# ADR 0012 — The main worktree is registrable but flagged "special"

**Status:** Accepted
**Date:** 2026-07-11

## Context
git treats the main checkout as a worktree, and its path also serves as the project identity (ADR 0003). We must decide whether it can be a worktree entry in its own right.

## Decision
- The **main worktree can be registered** as a normal worktree entry (e.g. `wt register --name main`).
- Such an entry is **flagged special/base** in both the stored structure and in `wt list` output, so consumers can tell the base checkout from linked worktrees.

## Consequences
- Uniform model: everything is a worktree entry; the base one just carries a flag.
- Useful concretely: an agent can register the entry points the **main repo** exposes, and later re-register them, distinguishing them from per-feature worktree entry points.
- `wt list` should visually mark the special entry; JSON output carries an explicit boolean/flag field.
