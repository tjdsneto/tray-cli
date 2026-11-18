# Desktop app

The Tray desktop app is a window for trays you own, plus the macOS menu bar.

## Launch

The app runs the `tray` binary on your `PATH` and talks to it on this machine only. Install the CLI first. Sign in with `tray login` before remote trays appear. Local trays work signed out.

`tray bar` still works. Running it at the same time as the app shows two menu bar icons.

## Window

On launch, placeholders stay up until the tray list and the attention list are both ready. Opening a tray later shows placeholders for that tray's rows only. A stalled remote call does not hold the window forever. The sidebar lists Attention, then local trays, directory and branch trays, remote trays you own, and trays you joined. A joined tray is listed by its local alias, with the owner's name under it. Attention is the badge set. A tray splits Pending, Open, and Completed, and the chips filter that list. Each row shows when it was created, its status, and the actions that apply. The open row is marked, and Close clears it. Copy ref copies `tray-ref item-id`. On a tray you joined, you see the items you added.

## Menu bar

The icon's menu lists trays and their open counts, grouped the same way as the sidebar. A tray opens the window on that tray. Open Tray shows the window. Copying a handoff line is the Copy button on an open item. `tray bar` still lists items and copies from its own menu.