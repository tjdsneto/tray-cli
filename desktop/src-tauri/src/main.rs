#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

use std::io::{BufRead, BufReader};
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

fn spawn_api() -> Result<ApiProcess, String> {
    let mut child = Command::new("tray")
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
    use tauri::{AppHandle, Manager};

    const TRAY_ICON: tauri::image::Image<'_> = tauri::include_image!("icons/icon.png");
    const REFRESH_INTERVAL: Duration = Duration::from_secs(30);
    const OPEN_ID: &str = "open";
    /// Clipboard lines are `tray-ref item-id` and never start with U+0001.
    const HEADER_PREFIX: char = '\u{1}';

    #[derive(Clone, serde::Deserialize)]
    struct Snapshot {
        badge: i64,
        #[serde(default, rename = "remoteStatus")]
        #[allow(dead_code)]
        remote_status: String,
        #[serde(default)]
        sections: Vec<Section>,
    }

    #[derive(Clone, serde::Deserialize)]
    struct Section {
        header: String,
        #[serde(default)]
        rows: Vec<Row>,
    }

    #[derive(Clone, serde::Deserialize)]
    struct Row {
        title: String,
        clipboard: String,
    }

    struct MenuEntry {
        id: String,
        label: String,
        enabled: bool,
    }

    #[derive(Debug)]
    enum MenuAction {
        OpenWindow,
        Copy(String),
        Ignore,
    }

    /// `&` is a menu mnemonic. Doubling it shows the character in the label.
    fn menu_label(title: &str) -> String {
        title.replace('&', "&&")
    }

    fn is_header_id(id: &str) -> bool {
        id.starts_with(HEADER_PREFIX)
    }

    fn menu_entries(snapshot: &Snapshot) -> Vec<MenuEntry> {
        let mut entries = vec![MenuEntry {
            id: OPEN_ID.to_string(),
            label: "Open".to_string(),
            enabled: true,
        }];
        for (section_index, section) in snapshot.sections.iter().enumerate() {
            entries.push(MenuEntry {
                id: format!("{HEADER_PREFIX}{section_index}"),
                label: menu_label(&section.header),
                enabled: false,
            });
            for row in &section.rows {
                entries.push(MenuEntry {
                    id: row.clipboard.clone(),
                    label: menu_label(&row.title),
                    enabled: true,
                });
            }
        }
        entries
    }

    /// The row id is the clipboard line, so a click still copies the row that
    /// was on the menu even if a refresh has already built a new one.
    fn menu_action(id: &str) -> MenuAction {
        if id == OPEN_ID {
            return MenuAction::OpenWindow;
        }
        if is_header_id(id) {
            return MenuAction::Ignore;
        }
        MenuAction::Copy(id.to_string())
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

    fn fetch_snapshot(base: &str, token: &str) -> Result<Snapshot, String> {
        let url = format!("{base}/v1/snapshot");
        let response = ureq::get(&url)
            .set("Authorization", &format!("Bearer {token}"))
            .call()
            .map_err(|err| err.to_string())?;
        let body = response.into_string().map_err(|err| err.to_string())?;
        serde_json::from_str(&body).map_err(|err| err.to_string())
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
            MenuAction::Copy(text) => {
                if let Ok(mut clipboard) = arboard::Clipboard::new() {
                    let _ = clipboard.set_text(text);
                }
            }
            MenuAction::Ignore => {}
        }
    }

    fn apply_snapshot(app: &AppHandle, snapshot: &Snapshot) {
        let entries = menu_entries(snapshot);
        let Some(menu) = build_menu(app, &entries) else {
            return;
        };
        let Some(tray) = app.tray_by_id("main") else {
            return;
        };
        if tray.set_menu(Some(menu)).is_err() {
            return;
        }
        let (title, tooltip) = badge_chrome(snapshot.badge);
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
        if stopped(app) {
            return;
        }
        let handle = app.clone();
        let _ = app.run_on_main_thread(move || apply_snapshot(&handle, &snapshot));
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
        let open_only = Snapshot {
            badge: 0,
            remote_status: String::new(),
            sections: Vec::new(),
        };
        let entries = menu_entries(&open_only);
        let Some(menu) = build_menu(app, &entries) else {
            eprintln!("tray menu: failed to build the initial menu");
            return Ok(());
        };
        let _tray = TrayIconBuilder::with_id("main")
            .icon(TRAY_ICON.clone())
            .icon_as_template(true)
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

        fn sample() -> Snapshot {
            Snapshot {
                badge: 2,
                remote_status: String::new(),
                sections: vec![Section {
                    header: "Today".into(),
                    rows: vec![
                        Row {
                            title: "Ship & tell".into(),
                            clipboard: "tray-a item-1".into(),
                        },
                        Row {
                            title: "Notes".into(),
                            clipboard: "tray-b item-2".into(),
                        },
                    ],
                }],
            }
        }

        #[test]
        fn menu_lists_open_then_disabled_headers_and_rows() {
            let entries = menu_entries(&sample());
            assert_eq!(entries[0].id, "open");
            assert!(entries[0].enabled);
            assert!(is_header_id(&entries[1].id));
            assert!(!entries[1].enabled);
            assert_eq!(entries[1].label, "Today");
            assert_eq!(entries[2].id, "tray-a item-1");
            assert_eq!(entries[2].label, "Ship && tell");
            assert_eq!(entries[3].id, "tray-b item-2");
            assert!(!entries[2].id.starts_with(HEADER_PREFIX));
        }

        #[test]
        fn open_and_headers_are_not_copied() {
            let entries = menu_entries(&sample());
            assert!(matches!(menu_action("open"), MenuAction::OpenWindow));
            assert!(matches!(menu_action(&entries[1].id), MenuAction::Ignore));
            match menu_action(&entries[2].id) {
                MenuAction::Copy(text) => assert_eq!(text, "tray-a item-1"),
                other => panic!("expected copy, got {other:?}"),
            }
        }

        #[test]
        fn click_copies_the_id_from_the_open_menu() {
            let clicked = menu_entries(&sample())[2].id.clone();
            let mut later = sample();
            later.sections[0].rows[0].clipboard = "tray-z item-9".into();
            let _refreshed = menu_entries(&later);
            match menu_action(&clicked) {
                MenuAction::Copy(text) => assert_eq!(text, "tray-a item-1"),
                other => panic!("expected copy, got {other:?}"),
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
        .invoke_handler(tauri::generate_handler![restart_api]);

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
    use super::{init_script, parse_handshake};

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
    fn parse_handshake_rejects_unusable_payloads() {
        assert!(parse_handshake("nope").is_err());
        assert!(parse_handshake(r#"{"port":0,"token":"abc"}"#).is_err());
        assert!(parse_handshake(r#"{"port":1,"token":""}"#).is_err());
    }
}
