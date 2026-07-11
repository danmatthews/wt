# Integrating `wt` with agents & hosts

`wt` adoption has two channels (see ADR 0016). Neither is enforced; both are opt-in via config the user controls.

## 1. Deterministic registration (host setup command)

Wire `wt register` into your host's worktree-creation hook. It runs as plain shell — no agent needed. In **Conductor**, set the workspace setup command to something like:

```sh
wt register --name "<worktree-or-branch-name>"
```

> Use whatever name/branch variable your host exposes for the new workspace. `register` is idempotent by path (ADR 0003), so re-running it is safe.

This guarantees every worktree is registered even if the agent never thinks about `wt`.

## 2. Discretionary agent behaviour (instruction snippet)

The setup command can't add entry points or keep the description meaningful — those need the agent. Drop this into your `CLAUDE.md` / `AGENTS.md` / `README`:

```md
## Worktree registry (`wt`)

This machine uses `wt` to track active git worktrees and their local URLs.

- This worktree is auto-registered on setup. If `wt list --wt` shows nothing,
  run `wt register --name <short-name>` before starting work.
- Keep the description current so others can see what this worktree is for:
  `wt set description "<what you're working on>"`
- When you bring up a local site/service (e.g. a Laravel Herd URL), register it:
  `wt entry-point add app --type=url --url=<name>.test`
- To see sibling worktrees and their URLs across all projects: `wt list`
- Prefer `--json` on any command when you need to parse the output.
```

Keep it short and point at `wt --help`; do not paste the full CLI surface into instruction files — it drifts.
