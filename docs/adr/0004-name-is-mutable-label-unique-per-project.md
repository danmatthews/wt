# ADR 0004 — `name` is a mutable label, unique within a project

**Status:** Accepted
**Date:** 2026-07-11

## Context
`name` is supplied at registration. It is not the identity (path is — ADR 0003), but it is how humans and agents refer to a worktree on the CLI.

## Decision
- `name` is a **mutable label**, not an identity.
- `name` need **not** be globally unique, but **must be unique within a single project** (base git repo).
- `name` serves as the **selector** for a worktree within a project on the CLI (e.g. `wt <project> entry-point add <name> ...`).

## Consequences
- Registering a second, different worktree path with a name already used in that project is a conflict → rejected.
- Since path is identity, a name can be changed on an existing entry without affecting its identity.
- Selection is two-level: pick a project, then a worktree by name.
- Open: how a project is referenced on the CLI (path vs. slug) — deferred to a later ADR.
