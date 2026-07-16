# `wt` — Design

A local, single-user registry of git worktrees, queryable by agents and tools (Conductor) to see what work is in flight per project, and where each worktree's running sites live. `wt` **records** worktrees created by other tools, and can also create (`wt create`) and delete (`wt remove`) worktrees itself (ADR 0002 amendment).

This document consolidates the decisions in `docs/adr/`. Where they disagree, the ADRs win.

## 1. Scope & non-goals

- **In:** a `~/.config/wt` registry, shared across concurrent processes on one macOS user account; registering/updating worktrees and their typed entry points; listing them.
- **Out (for now):** creating/removing git worktrees (ADR 0002); cross-machine/team sync (ADR 0001); URL derivation or any Herd awareness (ADR 0005); auth.

## 2. Domain model

```
Project (identity = main-worktree absolute path)
 ├─ display path (human-readable, stored)
 └─ Worktree entry (identity = worktree absolute path)   [1..n]
     ├─ name          friendly, unique within the project, CLI/list handle
     ├─ description?   free text
     ├─ special?       true iff this entry IS the project's main worktree
     └─ Entry point (identity = name, unique within the worktree)  [0..n]
         ├─ name
         ├─ type        e.g. "url"
         ├─ description?
         └─ <type fields>   url: { url: "mysite.test" }
```

Identities are **paths**, resolved from cwd via git plumbing — never typed on the CLI (ADR 0003, 0007). `name`s are mutable labels/selectors, not identities (ADR 0004, 0008).

## 3. Storage (ADR 0010)

- One TOML file **per project**: `~/.config/wt/projects/<sha256(main-worktree-path)[:16]>.toml`.
- The human path is stored inside the file. Concurrency: take the per-file `flock`, write temp, `rename` over the original.
- A project file is auto-created on first `register` of any of its worktrees.

Example `~/.config/wt/projects/1a2b3c4d5e6f7a8b.toml`:

```toml
project_path = "/Users/dan/code/acme"

[[worktree]]
path = "/Users/dan/code/acme"
name = "main"
description = "trunk"
special = true
registered_at = "2026-07-01T09:00:00Z"
updated_at = "2026-07-10T16:12:00Z"

  [[worktree.entry_point]]
  name = "app"
  type = "url"
  url  = "acme.test"
  description = "primary app"
  added_at = "2026-07-01T09:01:00Z"
  updated_at = "2026-07-01T09:01:00Z"

[[worktree]]
path = "/Users/dan/code/acme-feature-x"
name = "feature-x"
description = "checkout flow rework"
registered_at = "2026-07-09T11:30:00Z"
updated_at = "2026-07-11T14:05:00Z"

  [[worktree.entry_point]]
  name = "app"
  type = "url"
  url  = "feature-x.test"
  added_at = "2026-07-09T11:31:00Z"
  updated_at = "2026-07-09T11:31:00Z"

  [[worktree.entry_point]]
  name = "admin"
  type = "url"
  url  = "admin.feature-x.test"
  description = "back office"
  added_at = "2026-07-11T14:05:00Z"
  updated_at = "2026-07-11T14:05:00Z"
```

Timestamps are ISO-8601 UTC (ADR 0014); a worktree's `updated_at` bumps when any child entry point changes.

## 4. CLI surface

Every command supports `--json` (ADR 0013). All **mutating** commands self-identify from cwd and **fail if not inside a worktree** (ADR 0007). Reads are global (ADR 0009).

| Command | Scope | Purpose |
|---------|-------|---------|
| `wt create [--name <n>] [--path <p>] [--branch <b>] [--description <d>]` | cwd project | `git worktree add` on a new branch, then register it (ADR 0002 amendment). Path defaults to a sibling of the main worktree named after `--name`; branch defaults to `--name`. Prompts for a name if omitted (fails under `--json` / non-interactive stdin). |
| `wt register --name <n> [--description <d>]` | cwd worktree | Create/update this worktree's entry (idempotent by path). Auto-flags `special` when cwd is the main worktree. |
| `wt set name "<n>"` | cwd worktree | Rename this worktree. |
| `wt set description "<d>"` | cwd worktree | Re-describe this worktree. |
| `wt unregister` | cwd worktree | Remove this worktree's entry (does not touch disk). |
| `wt remove <name> [--force]` | cwd project | `git worktree remove` the named worktree, then unregister it (ADR 0002 amendment). Confirms interactively unless `--force`; refuses the main worktree and the current worktree; dirty trees need `--force`. |
| `wt entry-point add <ep> --type=url --url=<v> [--description <d>]` | cwd worktree | Attach an entry point. |
| `wt entry-point set <ep> [--name <new>] [--url <v>] [--description <d>]` | cwd worktree | Update an entry point. |
| `wt entry-point remove <ep>` | cwd worktree | Detach an entry point. |
| `wt list` | global | Print all projects → worktrees → entry points (tree); auto-prunes stale. |
| `wt list --wt` | cwd worktree | Print only the current worktree and its entry points. |

`wt ep` may alias `wt entry-point`.

## 5. JSON contract (ADR 0013, 0015)

Every command supports `--json` with a **consistent envelope**:

```json
{ "ok": true, "data": { … }, "pruned": [ { "name": "feature-x", "path": "…", "reason": "path_gone" } ] }
```
```json
{ "ok": false, "error": { "code": "not_in_worktree", "message": "…", "details": {} } }
```

- `pruned` (ADR 0006) rides alongside `data` when `wt list` deletes stale entries.
- `error.code` enum: `not_in_worktree`, `worktree_not_registered`, `worktree_not_found`, `name_conflict`, `entry_point_not_found`, `entry_point_name_conflict`, `unknown_entry_point_type`, `lock_timeout`, `io_error`, `usage`, `git_unavailable`, `git_error`.
- Failures exit `1` (single non-zero); agents branch on `error.code`, not exit status.

## 6. Lifecycle & staleness (ADR 0006)

- Well-behaved teardown: `wt unregister` drops the entry only; `wt remove <name>` also deletes the worktree from disk (ADR 0002 amendment).
- Safety net: because identity is the worktree path, any entry whose path no longer exists on disk is **stale**. `wt list` **deletes** it and **reports** it (human: a `pruned …` line; JSON: the `pruned` array). Pruning takes the project-file lock.
- `unregister` never deletes the worktree. `remove` does, but only after confirmation (or `--force`), and it deletes from git *before* unregistering so a failure leaves at worst a stale entry the prune above cleans up.

## 7. Open items to revisit later

- Future entry-point types beyond `url`.
- Whether a standalone `wt prune` is wanted in addition to `list`'s implicit prune.
- Versioning strategy for the `error.code` enum and JSON schema as they evolve.
