# wt app

A tiny macOS **windowed** app that shows the local [`wt`](../README.md) worktree
registry — the projects and worktrees you've registered — as a clean, live list.

It is **read-only**: it never creates, deletes, or edits registry entries. It
watches the registry folder and refreshes instantly, and pops a desktop
notification whenever a new worktree is registered.

## Stack

- **Tauri v2** — native shell, notifications, filesystem watching (Rust).
- **Vue 3 + Vite** — the UI, so the frontend stays plain and readable.
- **shadcn-vue** (reka-ui + Tailwind v4) — the component/design system.

## What it does

- Reads `~/.config/wt/projects/*.toml` (the location fixed by `wt`'s design) and
  groups it into **projects → worktrees → entry points**. Main worktrees are
  flagged; entry-point URLs are shown as badges. Clicking a URL opens it in the
  browser; clicking a worktree path reveals it in Finder.
- **Windowed:** a small, resizable window that opens centered on launch and shows
  a Dock icon; closing the window quits the app. Theme follows the system light/
  dark setting.
- **Live refresh:** a Rust `notify` watcher on the registry folder re-reads on any
  change and pushes the fresh list to the UI (`registry-changed` event).
- **Notifications:** when a worktree path appears that wasn't there on the previous
  read, it fires a native desktop notification naming the project and worktree.

## Develop

```sh
pnpm install
node scripts/generate-icons.mjs   # regenerate icons (only if you change the design)
pnpm tauri dev                     # run the app (vite + native shell)
```

## Build

```sh
pnpm tauri build                   # produces a .app / .dmg under src-tauri/target
```

## Where things live

| Path | What |
|------|------|
| `src/App.vue` | The list UI, initial load + `registry-changed` listener. |
| `src/components/WorktreeItem.vue` | One worktree row (name, path, entry points). |
| `src/components/ui/` | shadcn-vue components (card, badge, scroll-area, separator). |
| `src-tauri/src/registry.rs` | Reads & parses the TOML registry (unit-tested). |
| `src-tauri/src/lib.rs` | Window setup, file watcher, notifications. |
| `scripts/generate-icons.mjs` | Dependency-free generator for the app icon. |

## Tests

```sh
cargo test --manifest-path src-tauri/Cargo.toml   # registry parsing / diff logic
```
