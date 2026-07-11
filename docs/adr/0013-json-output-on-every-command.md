# ADR 0013 — Every command supports JSON output

**Status:** Accepted
**Date:** 2026-07-11

## Context
Agents and tools (Conductor) are first-class consumers of `wt`, not just humans. Parsing ASCII trees/tables is brittle.

## Decision
- **Every** `wt` command supports a `--json` flag emitting structured output (not just `list`).
- Default (no flag) output is human-friendly (trees/tables/confirmations); `--json` is the machine contract.
- Read commands emit their data as JSON; mutating commands emit a JSON result object (the resulting record / status).

## Consequences
- The JSON schema for worktree/entry-point/project records becomes a stable contract to version carefully.
- Agents never scrape human output.
- Errors under `--json` should also be machine-readable (structured error object + non-zero exit).
