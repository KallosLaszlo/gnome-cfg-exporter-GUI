<div align="center">

# GNOME Config Exporter

**A desktop application to backup and restore your entire GNOME desktop configuration.**

Built with [Go](https://go.dev/) + [Wails](https://wails.io/) + [Svelte](https://svelte.dev/) + [Tailwind CSS](https://tailwindcss.com/)

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![CodeQL](https://github.com/KallosLaszlo/gnome-cfg-exporter-GUI/actions/workflows/codeql.yml/badge.svg)](https://github.com/KallosLaszlo/gnome-cfg-exporter-GUI/actions/workflows/codeql.yml)
[![Support me on Ko-fi](https://img.shields.io/badge/Support%20me%20on-Ko--fi-F16061?logo=ko-fi&logoColor=white)](https://ko-fi.com/laszlokallos)

</div>

---

> **⚠️ This project is under active development.** Bug reports, feature requests, and pull requests are very welcome and greatly appreciated! Feel free to open an [issue](https://github.com/KallosLaszlo/gnome-cfg-exporter-GUI/issues) or submit a PR.

> **🐚 Prefer the command line?** Check out [GNOME Config Exporter (CLI)](https://github.com/KallosLaszlo/gnome-auto-steam-icon-fixer) — a standalone Bash script with the same functionality, no dependencies beyond `dconf` and `rsync`.

---

## Features

- **Full GNOME backup** — dconf database, extensions, themes, icons, fonts, wallpapers, GTK 3/4 settings, autostart, terminal profiles, monitor config, Nautilus scripts, default apps, online accounts, keyrings, desktop files, XDG user dirs
- **Selective backup & restore** — choose exactly which categories to include
- **Timeshift-inspired GUI** — clean, familiar desktop interface
- **CLI mode** — `--backup`, `--list`, `--help`, `--version` for scripting and automation
- **Scheduled backups** — hourly, daily, weekly, monthly via systemd user timers
- **Backup comparison** — diff two backups to see what changed
- **Safety backup** — automatic pre-restore safety snapshot of your current config
- **Version & user mismatch warnings** — alerts you before restoring across different GNOME versions or users
- **Fast startup** — backup metadata is cached for instant loading
- **Multi-distro support** — Arch, Fedora, Ubuntu, Debian, openSUSE, Void, Gentoo, NixOS

## Compatibility

**Fully compatible with the original [bash script](https://github.com/KallosLaszlo/gnome-cfg-exporter) version.** Backups created with the bash script are recognized and can be restored by this Go app, and vice versa. Both use the same directory structure, naming format (`YYYYMMDD_HHMMSS`), and metadata layout.

## Install

### Download from Releases (recommended)

1. Go to the [Releases](https://github.com/KallosLaszlo/gnome-cfg-exporter-GUI/releases) page
2. Download the latest `gnome-cfg-exporter` binary for your architecture
3. Make it executable and run:

```bash
chmod +x gnome-cfg-exporter
./gnome-cfg-exporter
```

Optionally, move it to your PATH:

```bash
sudo mv gnome-cfg-exporter /usr/local/bin/
```

### Build from Source

**Requirements:** Go 1.21+, Node.js 18+, npm, [Wails CLI v2](https://wails.io/docs/gettingstarted/installation)

```bash
# Install Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Clone and build
git clone https://github.com/KallosLaszlo/gnome-cfg-exporter-GUI.git
cd gnome-cfg-exporter
wails build
```

The binary will be at `build/bin/gnome-cfg-exporter`.

## Usage

### GUI

```bash
gnome-cfg-exporter
```

### CLI

```bash
gnome-cfg-exporter --backup      # Full backup
gnome-cfg-exporter --list        # List all backups
gnome-cfg-exporter --version     # Show version
gnome-cfg-exporter --help        # Show help
```

## Configuration

Config file: `~/.config/gnome-cfg-exporter/config.json`

Backups are stored in: `~/.local/share/gnome-cfg-exporter/` (configurable in Settings)

## Dependencies

| Command | Required | Description |
|---------|----------|-------------|
| `dconf` | Yes | GNOME settings database |
| `gsettings` | Yes | GSettings CLI |
| `rsync` | No | Fast file copying (falls back to native Go) |
| `gnome-extensions` | No | Extension management |
| `fc-cache` | No | Font cache rebuild after restore |

## Screenshots

*Coming soon*

## Contributing

This project is under active development. Contributions are welcome!

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/my-feature`)
3. Commit your changes (`git commit -am 'Add my feature'`)
4. Push to the branch (`git push origin feature/my-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.

## Author

**Kallos Laszlo** — [GitHub](https://github.com/KallosLaszlo) · [Ko-fi](https://ko-fi.com/laszlokallos)
