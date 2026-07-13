//! "Open with…" support: detect which supported apps are installed and open a
//! worktree path in the chosen one. macOS-native app set for now.

use std::path::{Path, PathBuf};
use std::process::Command;

use serde::Serialize;

#[derive(Debug, Clone, Serialize)]
pub struct Opener {
    pub id: String,
    pub label: String,
    pub available: bool,
}

struct AppSpec {
    id: &'static str,
    label: &'static str,
    /// Name passed to `open -a` (LaunchServices resolves it wherever it lives).
    app_name: &'static str,
    /// Candidate `.app` bundle locations to detect installation. `~` = home.
    bundles: &'static [&'static str],
}

/// Order here is the order shown in the menu.
const APPS: &[AppSpec] = &[
    AppSpec {
        id: "vscode",
        label: "VS Code",
        app_name: "Visual Studio Code",
        bundles: &[
            "/Applications/Visual Studio Code.app",
            "~/Applications/Visual Studio Code.app",
        ],
    },
    AppSpec {
        id: "cursor",
        label: "Cursor",
        app_name: "Cursor",
        bundles: &["/Applications/Cursor.app", "~/Applications/Cursor.app"],
    },
    AppSpec {
        id: "zed",
        label: "Zed",
        app_name: "Zed",
        bundles: &["/Applications/Zed.app", "~/Applications/Zed.app"],
    },
    AppSpec {
        id: "warp",
        label: "Warp",
        app_name: "Warp",
        bundles: &["/Applications/Warp.app", "~/Applications/Warp.app"],
    },
    AppSpec {
        id: "terminal",
        label: "Terminal",
        app_name: "Terminal",
        bundles: &[
            "/System/Applications/Utilities/Terminal.app",
            "/Applications/Utilities/Terminal.app",
        ],
    },
];

fn expand(bundle: &str) -> Option<PathBuf> {
    match bundle.strip_prefix("~/") {
        Some(rest) => dirs::home_dir().map(|h| h.join(rest)),
        None => Some(PathBuf::from(bundle)),
    }
}

fn is_available(spec: &AppSpec) -> bool {
    spec.bundles
        .iter()
        .filter_map(|b| expand(b))
        .any(|p| p.exists())
}

/// Every supported app with an `available` flag for the current machine.
pub fn list_openers() -> Vec<Opener> {
    APPS.iter()
        .map(|s| Opener {
            id: s.id.to_string(),
            label: s.label.to_string(),
            available: is_available(s),
        })
        .collect()
}

/// Open `path` in the app identified by `id` via macOS `open -a`.
pub fn open_in_app(id: &str, path: &str) -> Result<(), String> {
    let spec = APPS
        .iter()
        .find(|s| s.id == id)
        .ok_or_else(|| format!("Unknown application: {id}"))?;

    if !Path::new(path).exists() {
        return Err(format!("Path no longer exists: {path}"));
    }

    let status = Command::new("open")
        .args(["-a", spec.app_name, path])
        .status()
        .map_err(|e| e.to_string())?;

    if status.success() {
        Ok(())
    } else {
        Err(format!("Could not open {} with {}", path, spec.label))
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn lists_all_supported_apps_in_order() {
        let ids: Vec<_> = list_openers().into_iter().map(|o| o.id).collect();
        assert_eq!(ids, ["vscode", "cursor", "zed", "warp", "terminal"]);
    }

    #[test]
    fn rejects_unknown_app_id() {
        assert!(open_in_app("notepad", "/tmp").is_err());
    }
}
