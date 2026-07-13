# ADR 0010 — Human-readable file store with locking (no SQLite)

**Status:** Accepted
**Date:** 2026-07-11

## Context
Concurrent agents may write simultaneously. The store must be safe under concurrency but the user wants it human-readable/inspectable — SQLite is rejected in favour of plain files (JSON/YAML/TOML).

## Decision
- Backend is **plain-text files** under `~/.config/wt`, not SQLite.
- Concurrent writes are guarded by **locking** plus **atomic replace** (write temp → `rename`), so two agents can never half-write or clobber the same file.

Settled specifics:
- **Format: TOML** — most human-diffable, comments allowed, the `~/.config` idiom; avoids YAML's whitespace footguns and JSON's noise for a file agents rewrite.
- **Layout: one file per project** under `~/.config/wt/projects/`, so agents in different repos never contend for the same lock. Only two agents touching the *same* project serialize.
- **Filename: a hash of the main-worktree path** (first 16 hex of sha256). The human-readable path is stored *inside* the file. Hashes are stable, unique, and round-trip cleanly; slugs collide.
- **Lock granularity: per-file** (per-project), via advisory locking (e.g. `flock`) + atomic replace.

## Consequences
- Store is `cat`-able and diffable.
- Read-modify-write cycles take the project file's lock, write a temp file, then `rename` over the original.
- `~/.config/wt/projects/<hash>.toml` is the canonical unit of storage.
