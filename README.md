# wt

A local, single-user registry of git worktrees for agents and tools (e.g. Conductor). `wt` lets agents register the worktrees they're working on and the local URLs those worktrees expose, so `wt list` can show everything in flight across your projects.

`wt` **records** worktrees — it never creates or deletes them.

## Install

```sh
go build -o wt .
# move `wt` onto your PATH
```

## Quickstart

Run from inside a worktree — `wt` self-identifies the worktree and its project from the current directory.

```sh
wt register --name feature-x --description "checkout flow rework"
wt entry-point add app --type=url --url=feature-x.test
wt entry-point add admin --url=admin.feature-x.test --description "back office"
wt set description "checkout flow rework — now with Apple Pay"

wt list            # all projects → worktrees → entry points
wt list --wt       # just the current worktree
wt list --json     # machine-readable (every command supports --json)

wt unregister      # drop this worktree's entry (leaves the worktree on disk)
```

## How it works

- **Storage:** one TOML file per project under `~/.config/wt/projects/`, guarded by a per-project advisory lock and written atomically, so concurrent agents can't corrupt or clobber it.
- **Identity is by absolute path:** a project is keyed by its main-worktree path; a worktree by its own path. `register` is idempotent.
- **Reads are global, writes are cwd-scoped:** mutating commands operate on the worktree you're standing in and fail if you're not inside one.
- **Self-healing:** `wt list` prunes entries whose worktree has been deleted and reports what it removed.
- **JSON contract:** every command speaks `--json` with a `{ ok, data | error }` envelope; failures exit `1` with a stable `error.code`.

## Getting agents to use it

See [`docs/agent-integration.md`](docs/agent-integration.md) for the recommended Conductor setup command and a copy-paste `CLAUDE.md` / `AGENTS.md` snippet.

## Design docs

- [`docs/DESIGN.md`](docs/DESIGN.md) — consolidated design.
- [`docs/glossary.md`](docs/glossary.md) — ubiquitous language.
- [`docs/adr/`](docs/adr/) — the decision record (ADRs 0001–0016).
