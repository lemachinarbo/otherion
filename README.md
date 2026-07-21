# Otherion

A fork of [Aerion](https://github.com/hkdb/aerion) built specifically for **Omarchy Linux**.

---

## # Themes

- **Live Theme Switching via CLI & IPC**: Change theme dynamically on the fly with `otherion --theme-change <theme-name>`. Listens on Unix socket IPC to switch themes instantly without restarting the app.
- **Zero-Config Omarchy Integration**: Automatically registers with Omarchy Linux on startup, syncing window colors live whenever system themes change.
- **+15 Omarchy Color Schemes**: Pre-packaged with Catppuccin, Tokyo Night, Dracula, Nord, Flexoki, Rose Pine, Miasma, Lumon, Hackerman, and more.
- **Transparent Email Background**: Option to render email bodies with a transparent background that adapts text and link contrast to your active theme.
- **Contrast Avatars**: Dynamic text contrast for sender initial avatars.

---

## # Settings

- **VS Code-Style Full Window View**: Settings opens as a full-page view inside the main window via the left activity rail or sidebar gear icon, eliminating popup dialogs.
- **Action Toast Toggle**: Option under General Settings to disable success confirmation toasts (trashing/archiving) while keeping error warnings active.
- **Desktop Notification Toggle**: Option under General Settings to enable or disable new mail desktop notifications.
- **Toast Stack Limit**: Limits active toast displays to max 3 notifications.

---

## # Keyboard & Navigation

- **Auto-Preview Email Navigation**: Navigating the email list with keyboard shortcuts (`J`/`K` or arrow keys) automatically previews the selected email in the viewer pane without pressing `Enter`.
- **Shortcuts Manager**: Dedicated settings tab to search, view, and remap all pane navigation, composer, archive (`E`), spam, and sync shortcuts.
- **Window-Level Key Capture**: High-priority key recording supporting single keys (`E`), uppercase modifiers (`Shift+E`), or complex shortcuts (`Ctrl+K`) with `Esc` cancellation.

---

## # Appearance & UI

- **Transparent Sidebar**: Folder sidebar background set to `bg-background` to blend with custom Omarchy window themes.
- **Semantic Selection Overlays**: Active conversation selections and unread folder badges use theme tokens (`bg-muted`).

---

## # Installation

### Pre-built Release Binary
Download the latest binary from [GitHub Releases](https://github.com/lemachinarbo/otherion/releases):

```bash
curl -sSL https://github.com/lemachinarbo/otherion/releases/latest/download/otherion -o ~/.local/bin/otherion
chmod +x ~/.local/bin/otherion
```

### Build & Install from Source
```bash
git clone https://github.com/lemachinarbo/otherion.git
cd otherion
make install PREFIX=$HOME/.local
```

---

## License
Based on [Aerion](https://github.com/hkdb/aerion) by HKDB. Released under the Apache License 2.0.
