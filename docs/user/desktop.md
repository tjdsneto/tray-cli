# Desktop app

The Tray desktop app is a window for trays you own, plus the macOS menu bar.

## Launch

The app runs the `tray` binary on your `PATH` and talks to it on this machine only. Install the CLI first. Sign in with `tray login` before remote trays appear. Local trays work signed out.

`tray bar` still works. Running it at the same time as the app shows two menu bar icons.

## Window

Attention is the same set as the menu bar badge. A tray shows open work, newest first. Completed and archived items sit behind All. Opening a row loads its body. Buttons that do not apply to that item are not shown.

## Menu bar

The icon's menu still copies `tray-ref item-id` when you click a row. Open shows the window.