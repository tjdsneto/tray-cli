use std::fs;
use std::path::PathBuf;

fn main() {
    emit_app_version();
    // `allow-*` permissions are generated from this list. The capability file
    // references them; without that, Tauri rejects the invoke.
    tauri_build::try_build(
        tauri_build::Attributes::new().app_manifest(
            tauri_build::AppManifest::new().commands(&["restart_api", "app_version"]),
        ),
    )
    .expect("failed to run build script");
}

fn emit_app_version() {
    let manifest_dir = PathBuf::from(std::env::var("CARGO_MANIFEST_DIR").expect("CARGO_MANIFEST_DIR"));
    let conf = manifest_dir.join("tauri.conf.json");
    println!("cargo:rerun-if-changed={}", conf.display());
    let version = conf_version(&conf).unwrap_or_else(|| "0.0.0".into());
    println!("cargo:rustc-env=TRAY_APP_VERSION={version}");
}

fn conf_version(path: &std::path::Path) -> Option<String> {
    let text = fs::read_to_string(path).ok()?;
    let key = "\"version\"";
    let start = text.find(key)? + key.len();
    let rest = text[start..].trim_start().strip_prefix(':')?.trim_start();
    let rest = rest.strip_prefix('"')?;
    let end = rest.find('"')?;
    Some(rest[..end].to_string())
}
