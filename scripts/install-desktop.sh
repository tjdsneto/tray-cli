#!/usr/bin/env bash
# Replace /Applications/Tray.app with a new app bundle. Does not open a disk image.
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"

cd "${root}/desktop"
npx --yes @tauri-apps/cli@2 build --bundles app

osascript -e 'quit app "Tray"' >/dev/null 2>&1 || true
pkill -f 'Tray.app/Contents/MacOS/tray-desktop' >/dev/null 2>&1 || true
sleep 0.4
rm -rf /Applications/Tray.app
ditto "${root}/desktop/src-tauri/target/release/bundle/macos/Tray.app" /Applications/Tray.app
open -a /Applications/Tray.app
echo "Installed Tray at /Applications/Tray.app"
