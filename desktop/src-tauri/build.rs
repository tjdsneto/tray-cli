fn main() {
    // `allow-restart-api` is generated from this list. The capability file
    // references that permission; without it Tauri rejects the invoke.
    tauri_build::try_build(
        tauri_build::Attributes::new()
            .app_manifest(tauri_build::AppManifest::new().commands(&["restart_api"])),
    )
    .expect("failed to run build script");
}
