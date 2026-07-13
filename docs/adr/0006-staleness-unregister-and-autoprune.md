# ADR 0006 — Staleness: explicit unregister AND auto-prune

**Status:** Accepted
**Date:** 2026-07-11

## Context
Registries rot. When a worktree is deleted (by Conductor, git, or the user), its registry entry becomes stale. `wt` never deletes worktrees itself (ADR 0002), so it cannot rely on being told.

## Decision
Staleness is handled by **both**:
- (a) explicit teardown — `wt unregister` — for the well-behaved path; and
- (b) **auto-prune** — because identity is the worktree's absolute path (ADR 0003), an entry whose path no longer exists on disk is stale and is pruned automatically (at least on `wt list`).

Settled: auto-prune **deletes** the stale record (it does not merely hide it) and **notifies the user** that it was removed. On `wt list`, pruned entries are reported to the user (e.g. `pruned "feature-x" (path gone)`); under `--json`, they appear in a structured `pruned` collection in the output.

## Consequences
- `wt list` reflects reality even when tools forget to unregister, and self-heals by deleting dead records.
- `wt list` therefore has a write side-effect (deletion) — it takes the project file lock when it prunes.
- The user is always told what was removed, so a prune is never silent.
- Unregister removes only the registry record, never the worktree on disk.
