# ADR 0001 — Registry is a local, single-user datastore

**Status:** Accepted
**Date:** 2026-07-11

## Context
`wt` must be "shared" so multiple agents/tools on one machine see the same worktrees. The sharing boundary determines the entire storage design.

## Decision
The registry is **local to a single macOS user account**, stored under `~/.config/wt`. It is shared across processes/agents on that one machine, not across a team or network.

## Consequences
- No auth, no server, no sync layer. Just files on disk.
- "Shared" = concurrent local processes → the storage format must tolerate concurrent access (see forthcoming concurrency ADR).
- Herd URLs (inherently local, `*.test`) fit this boundary naturally.
- Cross-machine/team sharing is explicitly out of scope for now.
