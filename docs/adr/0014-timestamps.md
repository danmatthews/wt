# ADR 0014 — Per-record timestamps

**Status:** Accepted
**Date:** 2026-07-11

## Context
`wt list` benefits from recency ordering, and "when was this registered / last touched" is useful metadata for humans and agents triaging in-flight work.

## Decision
- Timestamps are **ISO-8601, UTC, second precision** (e.g. `2026-07-11T14:23:00Z`).
- **Worktree entry:** `registered_at` (set once at first register) and `updated_at` (bumped on any change to the entry or its entry points).
- **Entry point:** `added_at` and `updated_at`.
- `wt list` orders worktrees within a project by `updated_at` descending by default (most recently active first).

## Consequences
- Stored in the TOML per record; surfaced in `--json`.
- `updated_at` on a worktree bumps when a child entry point changes, so "most recent activity" is meaningful at the worktree level.
- Ordering is a display concern; `--json` consumers can re-sort as they like.
