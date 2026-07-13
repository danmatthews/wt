mod deps;
mod openers;
mod registry;

use std::collections::BTreeSet;
use std::sync::mpsc;
use std::time::Duration;

use notify::{RecursiveMode, Watcher};
use tauri::{AppHandle, Emitter, Manager, WindowEvent};
use tauri_plugin_notification::NotificationExt;

use registry::{read_registry, watch_root, worktree_paths, Project};

const WINDOW_LABEL: &str = "main";

/// Command backing the UI's initial load and its manual "refresh" button.
#[tauri::command]
fn list_projects() -> Vec<Project> {
    read_registry()
}

/// Command backing the Settings pane's environment status checks. Async +
/// `spawn_blocking` so the shell probes never block the main (UI) thread — the
/// pane opens instantly and shows its loader while this runs.
#[tauri::command]
async fn check_dependencies() -> Vec<deps::StatusCheck> {
    tauri::async_runtime::spawn_blocking(deps::check_dependencies)
        .await
        .unwrap_or_default()
}

/// The supported "Open with…" apps, each flagged available on this machine.
#[tauri::command]
fn list_openers() -> Vec<openers::Opener> {
    openers::list_openers()
}

/// Open a worktree path in the chosen app.
#[tauri::command]
async fn open_in_app(id: String, path: String) -> Result<(), String> {
    tauri::async_runtime::spawn_blocking(move || openers::open_in_app(&id, &path))
        .await
        .map_err(|e| e.to_string())?
}

pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_notification::init())
        .plugin(tauri_plugin_opener::init())
        .invoke_handler(tauri::generate_handler![
            list_projects,
            check_dependencies,
            list_openers,
            open_in_app
        ])
        .setup(|app| {
            let handle = app.handle().clone();

            // Single-window utility: closing the window quits the app.
            if let Some(win) = app.get_webview_window(WINDOW_LABEL) {
                let app_handle = handle.clone();
                win.on_window_event(move |event| {
                    if let WindowEvent::CloseRequested { .. } = event {
                        app_handle.exit(0);
                    }
                });
            }

            // Watch the registry for live refresh + new-worktree alerts.
            spawn_registry_watcher(handle);

            Ok(())
        })
        .run(tauri::generate_context!())
        .expect("error while running wt app");
}

/// Watches the registry directory and, on any change, re-reads it, pushes the
/// fresh list to the UI (`registry-changed`), and fires a desktop notification
/// for each worktree that appeared since the previous read.
fn spawn_registry_watcher(app: AppHandle) {
    std::thread::spawn(move || {
        let Some(root) = watch_root() else {
            return;
        };

        let (tx, rx) = mpsc::channel();
        let mut watcher = match notify::recommended_watcher(move |res| {
            let _ = tx.send(res);
        }) {
            Ok(w) => w,
            Err(_) => return,
        };
        if watcher.watch(&root, RecursiveMode::Recursive).is_err() {
            return;
        }

        // Seed with the current state so we don't alert for pre-existing entries.
        let mut known: BTreeSet<String> = worktree_paths(&read_registry());

        while let Ok(event) = rx.recv() {
            if event.is_err() {
                continue;
            }
            // Debounce: a single `wt` write can emit several fs events, and it
            // writes to a temp file then renames. Drain the burst before reading.
            while rx.recv_timeout(Duration::from_millis(250)).is_ok() {}

            let projects = read_registry();
            let current = worktree_paths(&projects);

            for path in current.difference(&known) {
                notify_new_worktree(&app, &projects, path);
            }

            known = current;
            let _ = app.emit("registry-changed", projects);
        }
    });
}

fn notify_new_worktree(app: &AppHandle, projects: &[Project], path: &str) {
    // Find the freshly-registered worktree so we can name it in the alert.
    let found = projects
        .iter()
        .find_map(|p| p.worktrees.iter().find(|w| w.path == path).map(|w| (p, w)));

    let (title, body) = match found {
        Some((project, wt)) => (
            "Worktree registered".to_string(),
            format!("{} · {}", project.display_name, wt.name),
        ),
        None => ("Worktree registered".to_string(), path.to_string()),
    };

    let _ = app
        .notification()
        .builder()
        .title(title)
        .body(body)
        .show();
}
