# ADR 0011 — Mutation command surface (`set`, entry-point verbs)

**Status:** Accepted
**Date:** 2026-07-11

## Context
Names and descriptions of both worktrees and entry points must be updatable. All mutations are cwd-scoped (ADR 0007).

## Decision
- Worktree fields: `wt set name "<new>"` and `wt set description "<new>"`, run from inside the worktree; **fail if not inside a worktree**.
- Entry points, from inside the worktree, use git-style sub-verbs under `entry-point`:
  - `wt entry-point add <ep-name> --type=url --url=<v> [--description "<d>"]`
  - `wt entry-point set <ep-name> [--name <new>] [--url <v>] [--description "<d>"]`
  - `wt entry-point remove <ep-name>`
- `register` remains idempotent-by-path (ADR 0003) and can set name/description at creation.

## Consequences
- `add`/`set`/`remove` mirror git's verbiage; `set` is a sub-verb of `entry-point`, not an overload of the top-level `wt set`.
- Entry points are addressed by their own name within the current worktree (ADR 0008).
- A short `wt ep …` alias may be provided for ergonomics.
