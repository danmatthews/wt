# ADR 0015 — JSON envelope and error contract

**Status:** Accepted
**Date:** 2026-07-11

## Context
Agents depend on `--json` output (ADR 0013). Success and failure need a predictable, uniform shape.

## Decision
- **Consistent envelope** on every `--json` response:
  - success: `{ "ok": true, "data": … }`
  - failure: `{ "ok": false, "error": … }`
  - Cross-cutting fields (e.g. the `pruned` list from ADR 0006) ride alongside `data`.
- **Error object:** `{ "code": "<enum>", "message": "<human>", "details": { … } }`.
- **`code` enum (starting set):** `not_in_worktree`, `worktree_not_registered`, `name_conflict`, `entry_point_not_found`, `entry_point_name_conflict`, `unknown_entry_point_type`, `lock_timeout`, `io_error`.
- **Exit codes:** a **single non-zero exit (`1`)** for all failures; agents branch on `error.code`, not the exit status.

## Consequences
- Agents check one field (`ok`) every time, then read `error.code` for specifics.
- The `code` enum is a versioned contract; new codes are additive.
- Human (non-`--json`) mode prints messages to stderr and still exits `1` on failure.
- Plain shell callers that want to branch must parse `error.code` (accepted; distinct per-class exit codes were rejected for simplicity).
