#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

use std::io::{BufRead, BufReader};
use std::path::{Path, PathBuf};
use std::process::{Child, ChildStdin, Command, Stdio};
use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::{mpsc, Mutex};
use std::thread;
use std::time::Duration;

use serde::Deserialize;
use tauri::{Manager, RunEvent, State};

const HANDSHAKE_TIMEOUT: Duration = Duration::from_secs(5);

/// Live `tray api` process. Stdin is held open so the child does not see EOF
/// and shut itself down. Dropping the value closes stdin, then kills the child.
struct ApiProcess {
    child: Child,
    stdin: Option<ChildStdin>,
    base: String,
    token: String,
}

impl Drop for ApiProcess {
    fn drop(&mut self) {
        drop(self.stdin.take());
        let _ = self.child.kill();
        let _ = self.child.wait();
    }
}

struct AppRuntime {
    api: Mutex<Option<ApiProcess>>,
    stop: AtomicBool,
}

#[derive(serde::Serialize)]
struct ApiEndpoint {
    base: String,
    token: String,
}

#[derive(Debug, Deserialize, PartialEq, Eq)]
struct Handshake {
    port: u16,
    token: String,
}

/// Assignment that runs with Tauri's initialization scripts, before the page parses HTML.
fn init_script(base: &str, token: &str) -> String {
    let value = serde_json::json!({
        "base": base,
        "token": token,
    });
    format!("window.__TRAY_API__ = {value};")
}

fn parse_handshake(line: &str) -> Result<Handshake, String> {
    let handshake: Handshake =
        serde_json::from_str(line.trim()).map_err(|err| format!("handshake: {err}"))?;
    if handshake.port == 0 {
        return Err("handshake port is 0".into());
    }
    if handshake.token.is_empty() {
        return Err("handshake token is empty".into());
    }
    Ok(handshake)
}

fn read_handshake_line(stdout: std::process::ChildStdout) -> Result<String, String> {
    let (tx, rx) = mpsc::channel();
    thread::spawn(move || {
        let mut reader = BufReader::new(stdout);
        let mut line = String::new();
        let read = reader.read_line(&mut line);
        let _ = tx.send(read.map(|_| line));
    });
    match rx.recv_timeout(HANDSHAKE_TIMEOUT) {
        Ok(Ok(line)) if !line.trim().is_empty() => Ok(line),
        Ok(Ok(_)) => Err("tray api closed stdout without a handshake".into()),
        Ok(Err(err)) => Err(err.to_string()),
        Err(mpsc::RecvTimeoutError::Timeout) => {
            Err("timed out waiting for tray api handshake".into())
        }
        Err(mpsc::RecvTimeoutError::Disconnected) => {
            Err("tray api handshake reader stopped".into())
        }
    }
}

fn tray_candidates(home: Option<&Path>) -> Vec<PathBuf> {
    let mut out = Vec::new();
    if let Some(home) = home {
        out.push(home.join("bin").join("tray"));
    }
    out.push(PathBuf::from("/opt/homebrew/bin/tray"));
    out.push(PathBuf::from("/usr/local/bin/tray"));
    out
}

fn resolve_tray(candidates: &[PathBuf]) -> Option<PathBuf> {
    candidates.iter().find(|path| path.is_file()).cloned()
}

fn tray_program() -> PathBuf {
    resolve_tray(&tray_candidates(std::env::var_os("HOME").as_deref().map(Path::new)))
        .unwrap_or_else(|| PathBuf::from("tray"))
}

fn spawn_api() -> Result<ApiProcess, String> {
    let mut child = Command::new(tray_program())
        .arg("api")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .spawn()
        .map_err(|err| err.to_string())?;

    let stdin = child.stdin.take();
    let Some(stdout) = child.stdout.take() else {
        drop(ApiProcess {
            child,
            stdin,
            base: String::new(),
            token: String::new(),
        });
        return Err("tray api stdout was not piped".into());
    };
    let line = match read_handshake_line(stdout) {
        Ok(line) => line,
        Err(err) => {
            // Drop closes stdin and kills the child we will not keep.
            drop(ApiProcess {
                child,
                stdin,
                base: String::new(),
                token: String::new(),
            });
            return Err(err);
        }
    };
    let handshake = match parse_handshake(&line) {
        Ok(handshake) => handshake,
        Err(err) => {
            drop(ApiProcess {
                child,
                stdin,
                base: String::new(),
                token: String::new(),
            });
            return Err(err);
        }
    };

    Ok(ApiProcess {
        child,
        stdin,
        base: format!("http://127.0.0.1:{}", handshake.port),
        token: handshake.token,
    })
}

#[tauri::command]
fn app_version() -> String {
    env!("TRAY_APP_VERSION").to_string()
}

#[tauri::command]
fn restart_api(state: State<AppRuntime>) -> Result<ApiEndpoint, String> {
    let mut guard = state.api.lock().map_err(|err| err.to_string())?;
    drop(guard.take());
    let next = spawn_api()?;
    let endpoint = ApiEndpoint {
        base: next.base.clone(),
        token: next.token.clone(),
    };
    *guard = Some(next);
    Ok(endpoint)
}

fn current_api(app: &tauri::AppHandle) -> Option<(String, String)> {
    let state = app.try_state::<AppRuntime>()?;
    let guard = state.api.lock().ok()?;
    let proc = guard.as_ref()?;
    Some((proc.base.clone(), proc.token.clone()))
}

fn shutdown_api(app: &tauri::AppHandle) {
    if let Some(state) = app.try_state::<AppRuntime>() {
        state.stop.store(true, Ordering::SeqCst);
        if let Ok(mut guard) = state.api.lock() {
            drop(guard.take());
        }
    }
}

#[cfg(target_os = "macos")]
mod menu_bar {
    use std::thread;
    use std::time::Duration;

    use super::{current_api, AppRuntime};
    use tauri::menu::{MenuBuilder, MenuEvent, MenuItemBuilder};
    use tauri::tray::TrayIconBuilder;
    use tauri::{AppHandle, Emitter, Manager};

    const TRAY_ICON: tauri::image::Image<'_> = tauri::include_image!("icons/trayTemplate.png");
    const REFRESH_INTERVAL: Duration = Duration::from_secs(30);
    const OPEN_ID: &str = "open";
    /// Clipboard lines are `tray-ref item-id` and never start with U+0001.
    const HEADER_PREFIX: char = '\u{1}';

    #[derive(Clone, serde::Deserialize)]
    struct Snapshot {
        badge: i64,
    }

    #[derive(Clone, serde::Deserialize)]
    struct ApiTray {
        #[serde(default)]
        kind: String,
        #[serde(rename = "ref")]
        tray_ref: String,
        name: String,
        #[serde(default)]
        detail: String,
        #[serde(default, rename = "openCount")]
        open_count: i64,
    }

    struct MenuEntry {
        id: String,
        label: String,
        enabled: bool,
    }

    #[derive(Debug)]
    enum MenuAction {
        OpenWindow,
        OpenTray(String),
        Ignore,
    }

    const TRAY_PREFIX: &str = "open-tray\u{1}";

    /// `&` is a menu mnemonic. Doubling it shows the character in the label.
    fn menu_name(tray: &ApiTray) -> &str {
        if tray.kind == "joined" && !tray.detail.trim().is_empty() {
            return tray.detail.trim();
        }
        tray.name.as_str()
    }

    fn menu_label(title: &str) -> String {
        title.replace('&', "&&")
    }

    fn is_header_id(id: &str) -> bool {
        id.starts_with(HEADER_PREFIX)
    }

    fn menu_entries(trays: &[ApiTray]) -> Vec<MenuEntry> {
        let mut entries = vec![MenuEntry {
            id: OPEN_ID.to_string(),
            label: "Open Tray".to_string(),
            enabled: true,
        }];
        const GROUPS: &[(&str, &[&str])] = &[
            ("Local", &["global"]),
            ("Directories", &["dir", "branch"]),
            ("Remote", &["remote"]),
            ("Joined", &["joined"]),
        ];
        for (label, kinds) in GROUPS {
            let mut group: Vec<&ApiTray> = trays
                .iter()
                .filter(|tray| kinds.contains(&tray.kind.as_str()))
                .collect();
            if group.is_empty() {
                continue;
            }
            group.sort_by(|a, b| menu_name(a).to_lowercase().cmp(&menu_name(b).to_lowercase()));
            entries.push(MenuEntry {
                id: format!("{HEADER_PREFIX}{label}"),
                label: (*label).to_string(),
                enabled: false,
            });
            for tray in group {
                entries.push(MenuEntry {
                    id: format!("{TRAY_PREFIX}{}", tray.tray_ref),
                    label: menu_label(&format!("{}    {}", menu_name(tray), tray.open_count)),
                    enabled: true,
                });
            }
        }
        entries
    }

    /// A tray row's id is the tray ref, so a click opens the tray that was on
    /// the menu even if a refresh has already built a new one.
    fn menu_action(id: &str) -> MenuAction {
        if id == OPEN_ID {
            return MenuAction::OpenWindow;
        }
        if is_header_id(id) {
            return MenuAction::Ignore;
        }
        if let Some(tray_ref) = id.strip_prefix(TRAY_PREFIX) {
            return MenuAction::OpenTray(tray_ref.to_string());
        }
        MenuAction::Ignore
    }

    /// Badge 0: tooltip "Tray" and no title. A positive badge uses the count for both.
    fn badge_chrome(badge: i64) -> (Option<String>, String) {
        if badge > 0 {
            let count = badge.to_string();
            (Some(count.clone()), count)
        } else {
            (None, "Tray".to_string())
        }
    }

    fn fetch_json<T: serde::de::DeserializeOwned>(url: &str, token: &str) -> Result<T, String> {
        let response = ureq::get(url)
            .set("Authorization", &format!("Bearer {token}"))
            .call()
            .map_err(|err| err.to_string())?;
        let body = response.into_string().map_err(|err| err.to_string())?;
        serde_json::from_str(&body).map_err(|err| err.to_string())
    }

    fn fetch_snapshot(base: &str, token: &str) -> Result<Snapshot, String> {
        fetch_json(&format!("{base}/v1/snapshot"), token)
    }

    fn fetch_trays(base: &str, token: &str) -> Result<Vec<ApiTray>, String> {
        fetch_json(&format!("{base}/v1/trays"), token)
    }

    fn show_main(app: &AppHandle) {
        if let Some(window) = app.get_webview_window("main") {
            let _ = window.show();
            let _ = window.set_focus();
        }
    }

    fn handle_menu(app: &AppHandle, event: MenuEvent) {
        match menu_action(event.id().as_ref()) {
            MenuAction::OpenWindow => show_main(app),
            MenuAction::OpenTray(tray_ref) => {
                show_main(app);
                let _ = app.emit("open-tray", tray_ref);
            }
            MenuAction::Ignore => {}
        }
    }

    fn apply_menu(app: &AppHandle, badge: i64, trays: &[ApiTray]) {
        let entries = menu_entries(trays);
        let Some(menu) = build_menu(app, &entries) else {
            return;
        };
        let Some(tray) = app.tray_by_id("main") else {
            return;
        };
        if tray.set_menu(Some(menu)).is_err() {
            return;
        }
        let (title, tooltip) = badge_chrome(badge);
        let _ = tray.set_title(title);
        let _ = tray.set_tooltip(Some(tooltip));
    }

    fn build_menu(app: &AppHandle, entries: &[MenuEntry]) -> Option<tauri::menu::Menu<tauri::Wry>> {
        let mut items = Vec::with_capacity(entries.len());
        for entry in entries {
            let item = MenuItemBuilder::with_id(&entry.id, &entry.label)
                .enabled(entry.enabled)
                .build(app)
                .ok()?;
            items.push(item);
        }
        let mut builder = MenuBuilder::new(app);
        for item in &items {
            builder = builder.item(item);
        }
        builder.build().ok()
    }

    fn stopped(app: &AppHandle) -> bool {
        app.try_state::<AppRuntime>()
            .map(|state| state.stop.load(std::sync::atomic::Ordering::SeqCst))
            .unwrap_or(true)
    }

    fn refresh_once(app: &AppHandle) {
        if stopped(app) {
            return;
        }
        let Some((base, token)) = current_api(app) else {
            return;
        };
        let Ok(snapshot) = fetch_snapshot(&base, &token) else {
            return;
        };
        let Ok(trays) = fetch_trays(&base, &token) else {
            return;
        };
        if stopped(app) {
            return;
        }
        let handle = app.clone();
        let badge = snapshot.badge;
        let _ = app.run_on_main_thread(move || apply_menu(&handle, badge, &trays));
    }

    fn refresh_loop(app: AppHandle) {
        loop {
            refresh_once(&app);
            for _ in 0..REFRESH_INTERVAL.as_secs() {
                if stopped(&app) {
                    return;
                }
                thread::sleep(Duration::from_secs(1));
            }
        }
    }

    pub fn start(app: &AppHandle) -> tauri::Result<()> {
        let entries = menu_entries(&[]);
        let Some(menu) = build_menu(app, &entries) else {
            eprintln!("tray menu: failed to build the initial menu");
            return Ok(());
        };
        let _tray = TrayIconBuilder::with_id("main")
            .icon(TRAY_ICON.clone())
            .icon_as_template(false)
            .menu(&menu)
            .tooltip("Tray")
            .show_menu_on_left_click(true)
            .on_menu_event(handle_menu)
            .build(app)?;

        let handle = app.clone();
        thread::spawn(move || refresh_loop(handle));
        Ok(())
    }

    #[cfg(test)]
    mod tests {
        use super::*;

        fn sample() -> Vec<ApiTray> {
            vec![
                ApiTray {
                    kind: "remote".into(),
                    tray_ref: "remote-1".into(),
                    name: "Inbox".into(),
                    detail: String::new(),
                    open_count: 21,
                },
                ApiTray {
                    kind: "dir".into(),
                    tray_ref: "dir:/proj".into(),
                    name: "Ship & tell".into(),
                    detail: String::new(),
                    open_count: 2,
                },
                ApiTray {
                    kind: "global".into(),
                    tray_ref: "notes".into(),
                    name: "notes".into(),
                    detail: String::new(),
                    open_count: 0,
                },
            ]
        }

        #[test]
        fn menu_lists_trays_by_kind_not_items() {
            let entries = menu_entries(&sample());
            assert_eq!(entries[0].id, "open");
            assert_eq!(entries[0].label, "Open Tray");
            assert!(is_header_id(&entries[1].id));
            assert_eq!(entries[1].label, "Local");
            assert!(!entries[1].enabled);
            assert_eq!(entries[2].id, format!("{TRAY_PREFIX}notes"));
            assert_eq!(entries[2].label, "notes    0");
            assert_eq!(entries[3].label, "Directories");
            assert_eq!(entries[4].label, "Ship && tell    2");
            assert_eq!(entries[5].label, "Remote");
            assert_eq!(entries[6].id, format!("{TRAY_PREFIX}remote-1"));
            assert!(entries.iter().all(|entry| !entry.label.contains("item")));
        }

        #[test]
        fn joined_menu_uses_the_alias() {
            let entries = menu_entries(&[ApiTray {
                kind: "joined".into(),
                tray_ref: "joined-1".into(),
                name: "lindris".into(),
                detail: "jessika-work".into(),
                open_count: 2,
            }]);
            assert_eq!(entries[2].label, "jessika-work    2");
        }

        #[test]
        fn tray_click_opens_that_tray() {
            let entries = menu_entries(&sample());
            assert!(matches!(menu_action("open"), MenuAction::OpenWindow));
            assert!(matches!(menu_action(&entries[1].id), MenuAction::Ignore));
            match menu_action(&entries[4].id) {
                MenuAction::OpenTray(tray_ref) => assert_eq!(tray_ref, "dir:/proj"),
                other => panic!("expected open tray, got {other:?}"),
            }
        }

        #[test]
        fn click_keeps_the_tray_from_the_open_menu() {
            let clicked = menu_entries(&sample())[4].id.clone();
            let later = vec![ApiTray {
                kind: "dir".into(),
                tray_ref: "dir:/other".into(),
                name: "other".into(),
                detail: String::new(),
                open_count: 1,
            }];
            let _refreshed = menu_entries(&later);
            match menu_action(&clicked) {
                MenuAction::OpenTray(tray_ref) => assert_eq!(tray_ref, "dir:/proj"),
                other => panic!("expected open tray, got {other:?}"),
            }
        }

        #[test]
        fn badge_zero_clears_the_title() {
            assert_eq!(badge_chrome(0), (None, "Tray".to_string()));
            assert_eq!(badge_chrome(-1), (None, "Tray".to_string()));
            assert_eq!(badge_chrome(3), (Some("3".into()), "3".to_string()));
        }
    }
}

#[cfg(not(target_os = "macos"))]
fn start_menu_bar(_app: &tauri::AppHandle) -> tauri::Result<()> {
    Ok(())
}

#[cfg(target_os = "macos")]
fn start_menu_bar(app: &tauri::AppHandle) -> tauri::Result<()> {
    menu_bar::start(app)
}

fn main() {
    let spawned = match spawn_api() {
        Ok(proc) => Some(proc),
        Err(err) => {
            eprintln!("tray api: {err}");
            None
        }
    };
    // Config windows are created before the setup hook, so this has to be
    // registered on the builder. It is injected with Tauri's other
    // initialization scripts: after `window` exists, before page scripts.
    let script = spawned
        .as_ref()
        .map(|proc| init_script(&proc.base, &proc.token));

    let mut builder = tauri::Builder::default()
        .plugin(tauri_plugin_single_instance::init(|app, _args, _cwd| {
            if let Some(window) = app.get_webview_window("main") {
                let _ = window.show();
                let _ = window.set_focus();
            }
        }))
        .manage(AppRuntime {
            api: Mutex::new(spawned),
            stop: AtomicBool::new(false),
        })
        .invoke_handler(tauri::generate_handler![restart_api, app_version]);

    if let Some(script) = script {
        builder = builder.append_invoke_initialization_script(script);
    }

    builder
        .setup(|app| {
            start_menu_bar(app.handle())?;
            Ok(())
        })
        .build(tauri::generate_context!())
        .expect("error while running tray desktop")
        .run(|app, event| {
            if let RunEvent::Exit = event {
                shutdown_api(app);
            }
        });
}

#[cfg(test)]
mod tests {
    use super::{init_script, parse_handshake, resolve_tray, tray_candidates};

    #[test]
    fn init_script_assigns_the_global_before_page_code() {
        let script = init_script("http://127.0.0.1:4312", "deadbeef");
        assert_eq!(
            script,
            r#"window.__TRAY_API__ = {"base":"http://127.0.0.1:4312","token":"deadbeef"};"#
        );
    }

    #[test]
    fn init_script_keeps_a_hostile_token_inside_the_string() {
        let token = "ab\"cd\\ef\n</script>";
        let script = init_script("http://127.0.0.1:9", token);
        let json = script
            .strip_prefix("window.__TRAY_API__ = ")
            .and_then(|rest| rest.strip_suffix(';'))
            .expect("assignment shape");
        let value: serde_json::Value = serde_json::from_str(json).expect("json stays intact");
        assert_eq!(value["base"], "http://127.0.0.1:9");
        assert_eq!(value["token"], token);
        assert!(!json.contains('\n'));
    }

    #[test]
    fn parse_handshake_reads_port_and_token() {
        let parsed = parse_handshake("{\"port\":4312,\"token\":\"abc\"}\n").unwrap();
        assert_eq!(parsed.port, 4312);
        assert_eq!(parsed.token, "abc");
    }

    #[test]
    fn resolve_tray_uses_the_first_binary_that_exists() {
        let dir = std::env::temp_dir().join(format!("tray-resolve-{}", std::process::id()));
        let _ = std::fs::remove_dir_all(&dir);
        std::fs::create_dir_all(dir.join("bin")).unwrap();
        let home_bin = dir.join("bin").join("tray");
        std::fs::write(&home_bin, b"").unwrap();
        let found = resolve_tray(&tray_candidates(Some(&dir))).unwrap();
        assert_eq!(found, home_bin);
        let _ = std::fs::remove_dir_all(&dir);
    }

    #[test]
    fn parse_handshake_rejects_unusable_payloads() {
        assert!(parse_handshake("nope").is_err());
        assert!(parse_handshake(r#"{"port":0,"token":"abc"}"#).is_err());
        assert!(parse_handshake(r#"{"port":1,"token":""}"#).is_err());
    }
}
