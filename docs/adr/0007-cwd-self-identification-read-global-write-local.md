# ADR 0007 — cwd self-identification; reads are global, writes are cwd-scoped

**Status:** Accepted
**Date:** 2026-07-11

## Context
`wt` runs inside a worktree an agent has just cd'd into. Earlier sketches passed an explicit `<project>` token on the CLI, which is painful (absolute paths) and redundant when the tool can self-identify.

## Decision
- `wt` **self-identifies from the current working directory** using git plumbing: it resolves the enclosing worktree (→ worktree-entry identity) and its main worktree (→ project identity). ADR 0003 identities are derived, never typed.
- **Reads are global:** `wt list` spans all projects regardless of cwd.
- **Writes are cwd-scoped:** every mutating command (`register`, `set …`, entry-point add/set/remove, `unregister`) operates on the worktree you are standing in and **fails if you are not inside a worktree**.
- `name` is the friendly handle shown in `wt list`; it is not needed to *target* mutations (cwd already does that).

## Consequences
- No `<project>`/`<name>` selector arguments on mutating commands — cwd is the selector.
- An agent orchestrating a worktree it is not inside cannot mutate it without cd'ing in first. Accepted for now.
- Open (grill): is the repo's **main worktree** itself "inside a worktree" for the purpose of these commands, i.e. is it registrable?
