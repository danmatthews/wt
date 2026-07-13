//! Reads the `wt` registry: one TOML file per project under
//! `~/.config/wt/projects/`, matching the schema in the project's DESIGN.md.
//! This module only ever *reads* — it never mutates the registry.

use std::collections::BTreeSet;
use std::path::{Path, PathBuf};

use serde::{Deserialize, Serialize};

// ---- On-disk shape (what we parse out of each TOML file) --------------------

#[derive(Debug, Deserialize)]
struct ProjectFile {
    project_path: String,
    #[serde(default)]
    worktree: Vec<WorktreeRaw>,
}

#[derive(Debug, Deserialize)]
struct WorktreeRaw {
    path: String,
    #[serde(default)]
    name: String,
    description: Option<String>,
    #[serde(default)]
    special: bool,
    registered_at: Option<String>,
    updated_at: Option<String>,
    #[serde(default, rename = "entry_point")]
    entry_point: Vec<EntryPointRaw>,
}

#[derive(Debug, Deserialize)]
struct EntryPointRaw {
    name: String,
    #[serde(rename = "type", default)]
    kind: String,
    description: Option<String>,
    url: Option<String>,
}

// ---- Shape we hand to the UI ------------------------------------------------

#[derive(Debug, Clone, Serialize)]
pub struct EntryPoint {
    pub name: String,
    #[serde(rename = "type")]
    pub kind: String,
    pub description: Option<String>,
    pub url: Option<String>,
}

#[derive(Debug, Clone, Serialize)]
pub struct Worktree {
    pub path: String,
    pub name: String,
    pub description: Option<String>,
    pub special: bool,
    pub registered_at: Option<String>,
    pub updated_at: Option<String>,
    pub entry_points: Vec<EntryPoint>,
}

#[derive(Debug, Clone, Serialize)]
pub struct Project {
    pub project_path: String,
    pub display_name: String,
    pub worktrees: Vec<Worktree>,
}

/// `~/.config/wt/projects` — the registry location fixed by wt's design
/// (note: NOT the OS config dir, which on macOS is Application Support).
pub fn registry_dir() -> Option<PathBuf> {
    dirs::home_dir().map(|h| h.join(".config").join("wt").join("projects"))
}

/// Sort key for "most recently touched": prefer `updated_at`, fall back to
/// `registered_at`. Empty when neither is set, sorting such entries last.
fn recency_key(w: &Worktree) -> &str {
    w.updated_at
        .as_deref()
        .or(w.registered_at.as_deref())
        .unwrap_or("")
}

fn display_name(project_path: &str) -> String {
    Path::new(project_path)
        .file_name()
        .and_then(|s| s.to_str())
        .map(|s| s.to_string())
        .unwrap_or_else(|| project_path.to_string())
}

/// Read and parse every project file in the registry, sorted for stable
/// display. Unreadable/unparseable files are skipped rather than failing the
/// whole read (the registry may be mid-write by another process).
pub fn read_registry() -> Vec<Project> {
    match registry_dir() {
        Some(dir) => read_registry_from(&dir),
        None => Vec::new(),
    }
}

/// Like [`read_registry`] but reads a caller-supplied directory (testable).
pub fn read_registry_from(dir: &Path) -> Vec<Project> {
    let Ok(entries) = std::fs::read_dir(dir) else {
        return Vec::new(); // dir may not exist yet — that's just "empty"
    };

    let mut projects = Vec::new();
    for entry in entries.flatten() {
        let path = entry.path();
        if path.extension().and_then(|e| e.to_str()) != Some("toml") {
            continue; // skip .lock and anything else
        }
        let Ok(text) = std::fs::read_to_string(&path) else {
            continue;
        };
        let Ok(file) = toml::from_str::<ProjectFile>(&text) else {
            continue;
        };

        let mut worktrees: Vec<Worktree> = file
            .worktree
            .into_iter()
            .map(|w| Worktree {
                path: w.path,
                name: w.name,
                description: w.description,
                special: w.special,
                registered_at: w.registered_at,
                updated_at: w.updated_at,
                entry_points: w
                    .entry_point
                    .into_iter()
                    .map(|e| EntryPoint {
                        name: e.name,
                        kind: e.kind,
                        description: e.description,
                        url: e.url,
                    })
                    .collect(),
            })
            .collect();

        // Most recently updated first. Timestamps are ISO-8601 UTC (Zulu), so
        // lexicographic string order matches chronological order. Fall back to
        // registered_at, then name, for entries missing an updated_at.
        worktrees.sort_by(|a, b| {
            recency_key(b).cmp(&recency_key(a)).then_with(|| {
                a.name.to_lowercase().cmp(&b.name.to_lowercase())
            })
        });

        projects.push(Project {
            display_name: display_name(&file.project_path),
            project_path: file.project_path,
            worktrees,
        });
    }

    // Most recently active project first. Worktrees are already sorted newest
    // first, so a project's recency is its first worktree's. Tie-break by name.
    projects.sort_by(|a, b| {
        project_recency(b).cmp(project_recency(a)).then_with(|| {
            a.display_name.to_lowercase().cmp(&b.display_name.to_lowercase())
        })
    });
    projects
}

/// A project's recency = the recency of its most-recently-updated worktree,
/// which (given worktrees are pre-sorted newest-first) is its first entry.
fn project_recency(p: &Project) -> &str {
    p.worktrees.first().map(recency_key).unwrap_or("")
}

/// The set of every registered worktree path — used to detect newly-added
/// worktrees between two reads (for notifications).
pub fn worktree_paths(projects: &[Project]) -> BTreeSet<String> {
    projects
        .iter()
        .flat_map(|p| p.worktrees.iter().map(|w| w.path.clone()))
        .collect()
}

/// Deepest ancestor of the registry dir that actually exists, so the watcher
/// can attach even before `wt` has created `~/.config/wt/projects`.
pub fn watch_root() -> Option<PathBuf> {
    let mut p = registry_dir()?;
    loop {
        if p.exists() {
            return Some(p);
        }
        match p.parent() {
            Some(parent) => p = parent.to_path_buf(),
            None => return None,
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    // The exact example from docs/DESIGN.md §3.
    const ACME: &str = r#"
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

  [[worktree.entry_point]]
  name = "admin"
  type = "url"
  url  = "admin.feature-x.test"
  description = "back office"
"#;

    fn tmpdir(label: &str) -> PathBuf {
        let dir =
            std::env::temp_dir().join(format!("wt-test-{}-{}", std::process::id(), label));
        let _ = std::fs::remove_dir_all(&dir);
        std::fs::create_dir_all(&dir).unwrap();
        dir
    }

    #[test]
    fn parses_and_sorts_the_design_example() {
        let dir = tmpdir("sorts");
        std::fs::write(dir.join("acme.toml"), ACME).unwrap();
        // A .lock sibling must be ignored.
        std::fs::write(dir.join("acme.toml.lock"), "").unwrap();

        let projects = read_registry_from(&dir);
        assert_eq!(projects.len(), 1);

        let p = &projects[0];
        assert_eq!(p.project_path, "/Users/dan/code/acme");
        assert_eq!(p.display_name, "acme"); // derived from the path's basename

        // Most recently updated sorts first: feature-x (07-11) beats main (07-10).
        assert_eq!(p.worktrees[0].name, "feature-x");
        assert_eq!(p.worktrees[1].name, "main");
        assert!(p.worktrees[1].special);

        // Entry points survive the round-trip, `type` included.
        let fx = &p.worktrees[0];
        assert_eq!(fx.entry_points.len(), 2);
        assert_eq!(fx.entry_points[0].kind, "url");
        assert_eq!(fx.entry_points[0].url.as_deref(), Some("feature-x.test"));

        std::fs::remove_dir_all(&dir).ok();
    }

    #[test]
    fn detects_newly_added_worktrees() {
        let before = read_registry_from(&tmpdir("added-before")); // empty
        let dir = tmpdir("added-after");
        std::fs::write(dir.join("acme.toml"), ACME).unwrap();
        let after = read_registry_from(&dir);

        let old = worktree_paths(&before);
        let new = worktree_paths(&after);
        let added: Vec<_> = new.difference(&old).collect();
        assert_eq!(added.len(), 2);

        std::fs::remove_dir_all(&dir).ok();
    }

    #[test]
    fn projects_sort_by_their_most_recent_worktree() {
        let dir = tmpdir("proj-order");
        // "zeta" has an older worktree than "acme" → acme should sort first,
        // proving the order is by activity, not alphabetical.
        std::fs::write(dir.join("acme.toml"), ACME).unwrap(); // newest: 2026-07-11
        std::fs::write(
            dir.join("zeta.toml"),
            r#"
project_path = "/Users/dan/code/zeta"
[[worktree]]
path = "/Users/dan/code/zeta"
name = "main"
updated_at = "2026-01-01T00:00:00Z"
"#,
        )
        .unwrap();

        let projects = read_registry_from(&dir);
        let names: Vec<_> = projects.iter().map(|p| p.display_name.as_str()).collect();
        assert_eq!(names, vec!["acme", "zeta"]); // by recency, not A→Z

        std::fs::remove_dir_all(&dir).ok();
    }

    #[test]
    fn missing_directory_reads_as_empty() {
        let missing = std::env::temp_dir().join("wt-does-not-exist-xyz");
        assert!(read_registry_from(&missing).is_empty());
    }
}
