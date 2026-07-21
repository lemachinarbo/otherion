# Otherion 🦖

A fork of [Aerion](https://github.com/hkdb/aerion) built specifically for **Omarchy Linux**.

**Why?** Because I wanted an email client that integrates nicely with Omarchy theming... and Aerion was still missing some small details to feel home.


<img width="1259" height="800" alt="screenshot-2026-07-20_23-38-44" src="https://github.com/user-attachments/assets/84d37a64-44ec-4483-9a11-eff6bcc3504a" />



---

## # Themes

- **Live Theme Switching**: Change theme dynamically on the fly with `otherion --theme-change <theme-name>`. Listens on Unix socket IPC to switch themes instantly without restarting the app. Automatically registers with Omarchy Linux on startup, syncing window colors live whenever system themes change.
- **+15 Omarchy Color Schemes**: Pre-packaged with Catppuccin, Tokyo Night, Dracula, Nord, Flexoki, Rose Pine, Miasma, Lumon, Hackerman, and more.
- **Transparent Email Background**: Option to render email bodies with a transparent background that adapts text and link contrast to your active theme.
- **Contrast Avatars**: Dynamic text contrast for sender initial avatars.

https://github.com/user-attachments/assets/a324257b-ee59-4fa4-9bf6-01b9603de5f2

---

## # Settings

- **Full Window View**: Settings opens as a full-page view inside the main window via the left activity rail or sidebar gear icon, eliminating popup dialogs.
- **Action Toast Toggle**: Option under General Settings to disable success confirmation toasts (trashing/archiving) while keeping error warnings active.
- **Desktop Notification Toggle**: Option under General Settings to enable or disable new mail desktop notifications.
- **Toast Stack Limit**: Limits active toast displays to max 3 notifications.

---

## # Keyboard & Navigation

- **Auto-Preview Email Navigation**: Navigating the email list with keyboard shortcuts (`J`/`K` or arrow keys) automatically previews the selected email in the viewer pane without pressing `Enter`.
- **Shortcuts Manager**: Dedicated settings tab to search, view, and remap all pane navigation, composer, archive (`E`), spam, and sync shortcuts.
- **Window-Level Key Capture**: High-priority key recording supporting single keys (`E`), uppercase modifiers (`Shift+E`), or complex shortcuts (`Ctrl+K`) with `Esc` cancellation.

<img width="1396" height="982" alt="screenshot-2026-07-20_23-33-12" src="https://github.com/user-attachments/assets/17109fbd-2262-4da1-adff-e5cd5cf4d28c" />


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
