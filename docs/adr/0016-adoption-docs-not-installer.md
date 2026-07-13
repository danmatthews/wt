# ADR 0016 — Adoption via docs, not a skill installer (for now)

**Status:** Accepted
**Date:** 2026-07-11

## Context
We want agents/tools to actually use `wt`. Options considered: a `wt skill` installer that writes host-specific SKILL.md files, an MCP server, and per-host instruction snippets. A key constraint surfaced: host worktree-creation hooks (e.g. Conductor's setup command) run as **plain shell, not through the agent** — so they can register deterministically but cannot make the agent add entry points or maintain descriptions.

## Decision
Split adoption and keep v1 minimal:
- **Deterministic registration** rides the host's setup shell command, e.g. Conductor runs `wt register --name <name>` when it creates a worktree. No agent involvement.
- **Discretionary agent behaviour** (entry points, description upkeep, consulting `wt list`) is prompted by a short **section the user adds to their README / `CLAUDE.md` / `AGENTS.md`**. `wt` ships a recommended snippet (see `docs/agent-integration.md`).
- **Deferred:** a `wt skill install` host-specific installer, and an MCP server. Revisit if a host makes prose instructions unreliable.

## Consequences
- Nothing host-specific to build or maintain in v1 — focus goes to the tool itself.
- Registration coverage depends on the user wiring the setup command; entry-point/description upkeep depends on the user adding the instruction snippet. Both are documented, neither is enforced.
- `wt skill --print` / MCP remain clean future additions layered on the same CLI.
- Polyscope's instruction/skill format is unconfirmed and intentionally not targeted yet.
