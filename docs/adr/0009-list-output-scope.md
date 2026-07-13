# ADR 0009 — `wt list` scope: full tree by default, `--wt` for current worktree

**Status:** Accepted
**Date:** 2026-07-11

## Context
`wt list` is consumed by both humans and agents. Scope and format are a contract.

## Decision
- **Default:** `wt list` prints **all** registered projects → their worktrees → each worktree's entry points (a three-level tree).
- **`wt list --wt`:** prints **only the current worktree** and its entry points, when cwd is inside a registered worktree.
- Auto-prunes stale entries as it goes (ADR 0006).
- Supports `--json` for machine consumption (see ADR 0013).

## Consequences
- Default output is global; it does not depend on cwd.
- `--wt` requires cwd to be inside a worktree (mirrors the write-scope rule of ADR 0007).
