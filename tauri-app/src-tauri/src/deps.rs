//! Environment checks shown in the Settings pane: are the external tools that
//! `wt` integrates with actually available on this machine?

use std::path::Path;
use std::process::Command;

use serde::Serialize;

#[derive(Debug, Clone, Serialize)]
pub struct StatusCheck {
    pub id: String,
    pub label: String,
    pub ok: bool,
    pub detail: String,
}

/// Resolve a command the way the user's terminal would. A GUI app inherits a
/// minimal PATH (typically just `/usr/bin:/bin:/usr/sbin:/sbin`), so tools under
/// Homebrew, `~/go/bin`, Herd, etc. are invisible to `std::env::var("PATH")`.
/// Asking a login + interactive shell reflects what the user actually has.
fn resolve_command(cmd: &str) -> Option<String> {
    let shell = std::env::var("SHELL").unwrap_or_else(|_| "/bin/zsh".to_string());
    let output = Command::new(&shell)
        .args(["-lic", &format!("command -v -- {cmd} 2>/dev/null")])
        .output()
        .ok()?;
    if !output.status.success() {
        return None;
    }
    let path = String::from_utf8_lossy(&output.stdout)
        .lines()
        .next()
        .unwrap_or("")
        .trim()
        .to_string();
    (!path.is_empty()).then_some(path)
}

/// Laravel Herd ships as a macOS app bundle; note it even when the CLI is missing.
fn herd_app() -> Option<String> {
    ["/Applications/Herd.app", "/Applications/Laravel Herd.app"]
        .into_iter()
        .find(|p| Path::new(p).exists())
        .map(str::to_string)
}

/// The status checks, in display order. Each probe spawns a shell (which sources
/// the user's rc files and is slow), so run them concurrently rather than
/// back-to-back.
pub fn check_dependencies() -> Vec<StatusCheck> {
    let wt_probe = std::thread::spawn(|| resolve_command("wt"));
    let herd_probe = std::thread::spawn(|| resolve_command("herd"));
    let wt = wt_probe.join().unwrap_or(None);
    let herd_cli = herd_probe.join().unwrap_or(None);

    vec![
        StatusCheck {
            id: "wt".to_string(),
            label: "wt CLI".to_string(),
            ok: wt.is_some(),
            detail: wt.unwrap_or_else(|| "Not found on your PATH".to_string()),
        },
        StatusCheck {
            id: "herd".to_string(),
            label: "Laravel Herd".to_string(),
            ok: herd_cli.is_some(),
            detail: match (herd_cli, herd_app()) {
                (Some(path), _) => format!("herd CLI at {path}"),
                (None, Some(app)) => {
                    format!("App installed ({app}), but `herd` CLI is not on your PATH")
                }
                (None, None) => "Not found on your PATH".to_string(),
            },
        },
    ]
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn returns_the_two_expected_checks_in_order() {
        // Result count/ids are stable regardless of what's installed.
        let checks = check_dependencies();
        let ids: Vec<_> = checks.iter().map(|c| c.id.as_str()).collect();
        assert_eq!(ids, ["wt", "herd"]);
        assert_eq!(checks[0].label, "wt CLI");
        assert_eq!(checks[1].label, "Laravel Herd");
    }
}
