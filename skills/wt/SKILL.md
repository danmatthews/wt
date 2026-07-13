---
name: wt
description: Use when working inside a git worktree on a machine that tracks active worktrees with the `wt` CLI — to register the current worktree, keep its description current, and record the local URLs/services it exposes so sibling worktrees are discoverable. Also use to discover other active worktrees and their URLs across projects (`wt list`). Triggers when you start work in a fresh worktree, bring up a local site/service (e.g. a `.test` URL), or want to see what else is in flight.
---

# wt — git worktree registry

`wt` is a local, single-user registry of git worktrees and the local URLs they expose. It **records** worktrees; it never creates or deletes them. Use it to make the worktree you're in — and any services it brings up — visible to other agents and tools on this machine.

## Prerequisite: check `wt` is installed

`wt` is a separate CLI binary; this skill only tells you how to use it. Before running any `wt` command:

```sh
command -v wt >/dev/null 2>&1 || echo "wt not on PATH"
```

If `wt` is not on PATH, **do not attempt to install it and do not block your task** — `wt` is optional. Silently skip the registry steps below (optionally mention once that `wt` isn't installed, with a pointer to build it via `go build -o wt .` from the wt repo). Everything below assumes `wt` is available.

## Working in a worktree

Run `wt` from inside the worktree — it self-identifies the worktree and its project from the current directory. Mutating commands operate on the worktree you're standing in and fail if you're not inside one.

- **Confirm this worktree is registered.** It's often auto-registered by the host's setup command. If `wt list --wt` shows nothing, register it:
  ```sh
  wt register --name <short-name>
  ```
  `register` is idempotent by path, so re-running it is safe.

- **Keep the description current** so others can see what this worktree is for:
  ```sh
  wt set description "<what you're working on>"
  ```

- **Record local URLs/services** when you bring one up (e.g. a Laravel Herd `.test` URL, a dev server):
  ```sh
  wt entry-point add app --type=url --url=<name>.test
  wt entry-point add admin --url=admin.<name>.test --description "back office"
  ```

- **See sibling worktrees and their URLs** across all projects:
  ```sh
  wt list
  ```

- **Drop this worktree's entry** when you're done with it (leaves the worktree on disk):
  ```sh
  wt unregister
  ```

## Tips

- Every command supports `--json`, emitting a `{ ok, data | error }` envelope; failures exit `1` with a stable `error.code`. Prefer `--json` when you need to parse output.
- `wt list` is self-healing: it prunes entries whose worktree has been deleted and reports what it removed.
- Don't memorize the full CLI surface — run `wt --help` (or `wt <command> --help`) to confirm flags before relying on them.
