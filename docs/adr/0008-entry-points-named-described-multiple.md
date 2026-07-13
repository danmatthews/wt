# ADR 0008 — Entry points are named, described, and can be many per worktree

**Status:** Accepted
**Date:** 2026-07-11

## Context
Real projects expose several access points (app, admin, API, Vite dev server). Users need to see what each is for, and to update/remove individual ones.

## Decision
An entry point is a record of shape:

```
{ name, type, description?, <type-specific fields> }
```

- A worktree may have **many** entry points, including several of the **same type**.
- Each entry point has its **own name** (unique within its worktree) — the handle used to update or remove it.
- Each entry point may carry its own **description** ("what it's for / does").
- Type `url` adds a `url` field (e.g. `mysite.test`).

## Consequences
- Entry-point identity = (worktree, entry-point name). `--type` is not the key; multiple URLs coexist.
- Command surface must cover add / set (rename, re-describe, change value) / remove — see ADR 0011.
- The record is extensible to future types without schema churn.
