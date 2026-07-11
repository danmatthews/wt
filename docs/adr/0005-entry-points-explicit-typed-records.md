# ADR 0005 — Entry points are explicit, typed records; URLs are provided, not derived

**Status:** Accepted
**Date:** 2026-07-11

## Context
A worktree can expose ways to be accessed. The first type is a URL (e.g. a Laravel Herd site at `name.test`). We must decide whether `wt` infers URLs (Herd convention) or treats them as dumb data.

## Decision
- Entry points are **explicit, typed records** attached to a worktree.
- The first type is `url`.
- URLs are **provided by the caller**, not derived from the worktree/name. `wt` has **no knowledge of Herd** or any URL convention.
- Managed via a dedicated command, shape approximately:
  `wt <project> entry-point add <name> --type=url --url=mysite.test`

## Consequences
- The type system is extensible (future types beyond `url`) without changing the core worktree model.
- `wt` stays decoupled from Herd; whoever knows the URL passes it in.
- Open questions (to grill): does an entry point have its own label? Can a worktree have multiple URLs? How is an individual entry point removed?
