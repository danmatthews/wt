# wt

A local, single-user registry of git worktrees for agents and tools (e.g. Conductor). `wt` lets agents register the worktrees they're working on and the local URLs those worktrees expose, so `wt list` can show everything in flight across your projects.

`wt` **records** worktrees created by other tools — and can also create and tear down worktrees for you in a single step with `wt create` / `wt remove`.

## Install

```sh
go build -o wt .
# move `wt` onto your PATH
```

## Quickstart

Run from inside a worktree — `wt` self-identifies the worktree and its project from the current directory.

```sh
# create a worktree and register it in one step (prompts for a name if omitted)
wt create --name feature-x --description "checkout flow rework"

# ...or register a worktree that already exists on disk
wt register --name feature-x --description "checkout flow rework"
wt entry-point add app --type=url --url=feature-x.test
wt entry-point add admin --url=admin.feature-x.test --description "back office"
wt set description "checkout flow rework — now with Apple Pay"

wt list            # all projects → worktrees → entry points
wt list --wt       # just the current worktree
wt list --json     # machine-readable (every command supports --json)

wt unregister      # drop this worktree's entry (leaves the worktree on disk)
wt remove feature-x  # delete the worktree and its entry (run from another worktree; --force to skip the prompt)
```

## Creating and removing worktrees

`wt` normally just records worktrees that already exist. Two commands go further and drive git for you (the only place `wt` mutates git — see [ADR 0002](docs/adr/0002-records-only-never-manages-git.md)).

### `wt create`

Runs `git worktree add` on a **new branch**, then registers the result. Run it from anywhere inside the project.

```sh
wt create --name feature-x --description "checkout flow rework"
wt create                 # no --name → prompts for one
```

| Flag | Default | Meaning |
|------|---------|---------|
| `--name` | *(prompted)* | Friendly name for the worktree. Required under `--json` (can't prompt). |
| `--path` | sibling of the main worktree, named after `--name` | Where to create the worktree. |
| `--branch` | `--name` | Name of the new branch to check out. |
| `--description` | — | What the worktree is for. |
| `--app` | — | Application working on it, e.g. `"Conductor.app"`. |

To adopt a worktree that **already exists** (or check out an existing branch), create it with `git` and use `wt register` instead.

### `wt remove`

Runs `git worktree remove` on a named worktree, then unregisters it — deleting it from disk. Target it **by name** and run from a *different* worktree (git won't remove the one you're standing in).

```sh
wt remove feature-x           # asks: remove worktree "feature-x" at …? [y/N]
wt remove feature-x --force   # skip the prompt; also removes a dirty worktree
```

- Prompts for confirmation unless `--force`; refuses without `--force` under `--json` or a non-interactive stdin.
- A worktree with uncommitted or untracked changes needs `--force` (git's own protection).
- Refuses the project's main worktree and the worktree you're currently in.
- Removes from git *before* unregistering, so an interrupted removal leaves at worst a stale entry that `wt list` prunes — never an untracked live worktree.

Use `wt unregister` instead if you only want to stop tracking a worktree without deleting it.

## How it works

- **Storage:** one TOML file per project under `~/.config/wt/projects/`, guarded by a per-project advisory lock and written atomically, so concurrent agents can't corrupt or clobber it.
- **Identity is by absolute path:** a project is keyed by its main-worktree path; a worktree by its own path. `register` is idempotent.
- **Reads are global, writes are cwd-scoped:** mutating commands operate on the worktree you're standing in and fail if you're not inside one. `wt create` and `wt remove` are the exceptions that also *touch git* — a single `git worktree add` / `git worktree remove` — around the registry update; `remove` targets a sibling worktree by name and confirms (or takes `--force`) before deleting (see ADR 0002).
- **Self-healing:** `wt list` prunes entries whose worktree has been deleted and reports what it removed.
- **JSON contract:** every command speaks `--json` with a `{ ok, data | error }` envelope; failures exit `1` with a stable `error.code`.

## Getting agents to use it

See [`docs/agent-integration.md`](docs/agent-integration.md) for the recommended Conductor setup command and a copy-paste `CLAUDE.md` / `AGENTS.md` snippet.

### Install as an agent skill

If you use the [Skills CLI](https://github.com/vercel-labs/skills), install `wt`'s agent instructions as a skill:

```sh
pnpx skills add danmatthews/wt   # or: npx skills add danmatthews/wt
```

This installs [`skills/wt/SKILL.md`](skills/wt/SKILL.md) into your agent (Claude Code, Cursor, etc.) so it knows how and when to use `wt`. The skill carries only the instructions — it does not install the `wt` binary, so make sure `wt` is on your PATH (see [Install](#install)).

## Design docs

- [`docs/DESIGN.md`](docs/DESIGN.md) — consolidated design.
- [`docs/glossary.md`](docs/glossary.md) — ubiquitous language.
- [`docs/adr/`](docs/adr/) — the decision record (ADRs 0001–0016).
