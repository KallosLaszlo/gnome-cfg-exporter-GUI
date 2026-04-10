package backup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gnome-cfg-exporter/internal/dconf"
	"gnome-cfg-exporter/internal/distro"
	"gnome-cfg-exporter/internal/fsutil"
)

// Category represents a backup category.
type Category struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Desc string `json:"desc"`
}

// AllCategories returns the list of all available backup categories.
func AllCategories() []Category {
	return []Category{
		{ID: "metadata", Name: "Metadata", Desc: "System info, timestamps, GNOME version"},
		{ID: "dconf", Name: "dconf Database", Desc: "Full + selective section dumps"},
		{ID: "extensions", Name: "Extensions", Desc: "GNOME Shell extension files"},
		{ID: "themes", Name: "Themes", Desc: "GTK/Shell themes"},
		{ID: "icons", Name: "Icons", Desc: "Icon themes"},
		{ID: "fonts", Name: "Fonts", Desc: "User fonts"},
		{ID: "wallpapers", Name: "Wallpapers", Desc: "Background images"},
		{ID: "gtk", Name: "GTK 3/4 Settings", Desc: "settings.ini, gtk.css, bookmarks"},
		{ID: "autostart", Name: "Autostart", Desc: "Startup applications"},
		{ID: "terminal", Name: "Terminal Profiles", Desc: "GNOME Terminal/Console profiles (dconf)"},
		{ID: "monitors", Name: "Monitor Config", Desc: "monitors.xml display layout"},
		{ID: "nautilus", Name: "Nautilus", Desc: "File manager scripts & prefs"},
		{ID: "mimeapps", Name: "Default Apps", Desc: "MIME type associations"},
		{ID: "goa", Name: "Online Accounts", Desc: "GNOME Online Accounts"},
		{ID: "keyrings", Name: "Keyrings", Desc: "Passwords, tokens, SSH keys"},
		{ID: "desktop_files", Name: "Desktop Files", Desc: "Custom .desktop launchers"},
		{ID: "xdg_dirs", Name: "XDG User Dirs", Desc: "Desktop, Documents, Downloads paths"},
	}
}

// Config holds options for a backup run.
type Config struct {
	Categories  []string `json:"categories"`
	Description string   `json:"description"`
	BackupDir   string   `json:"backupDir"`
}

// Result holds the outcome of a backup.
type Result struct {
	Path      string   `json:"path"`
	Timestamp string   `json:"timestamp"`
	Size      string   `json:"size"`
	FileCount int      `json:"fileCount"`
	Errors    []string `json:"errors"`
}

// ProgressFunc is called with progress percentage and status message.
type ProgressFunc func(pct int, msg string)

// Run executes a backup with the given config.
func Run(cfg Config, progress ProgressFunc) (*Result, error) {
	home, _ := os.UserHomeDir()
	ts := time.Now().Format("20060102_150405")
	bdir := filepath.Join(cfg.BackupDir, ts)

	if err := os.MkdirAll(bdir, 0755); err != nil {
		return nil, fmt.Errorf("create backup dir: %w", err)
	}

	result := &Result{Path: bdir, Timestamp: ts}

	catSet := make(map[string]bool)
	for _, c := range cfg.Categories {
		catSet[c] = true
	}

	total := len(cfg.Categories)
	if total == 0 {
		total = 1
	}

	for i, cat := range cfg.Categories {
		pct := (i + 1) * 100 / total
		var err error

		switch cat {
		case "metadata":
			progress(pct, "Saving metadata...")
			err = backupMetadata(bdir, home)
		case "dconf":
			progress(pct, "Dumping dconf database...")
			err = backupDconf(bdir)
		case "extensions":
			progress(pct, "Saving extensions...")
			err = backupDir(filepath.Join(home, ".local/share/gnome-shell/extensions"), filepath.Join(bdir, "extensions"))
		case "themes":
			progress(pct, "Saving themes...")
			err = backupMultipleDirs(home, bdir, map[string]string{
				"themes_local":  ".local/share/themes",
				"themes_legacy": ".themes",
			})
		case "icons":
			progress(pct, "Saving icons...")
			err = backupMultipleDirs(home, bdir, map[string]string{
				"icons_local":  ".local/share/icons",
				"icons_legacy": ".icons",
			})
		case "fonts":
			progress(pct, "Saving fonts...")
			err = backupMultipleDirs(home, bdir, map[string]string{
				"fonts_local":  ".local/share/fonts",
				"fonts_legacy": ".fonts",
			})
		case "wallpapers":
			progress(pct, "Saving wallpapers...")
			err = backupDir(filepath.Join(home, ".local/share/backgrounds"), filepath.Join(bdir, "files/backgrounds"))
		case "gtk":
			progress(pct, "Saving GTK settings...")
			err = backupMultipleDirs(home, bdir, map[string]string{
				"gtk-3.0": ".config/gtk-3.0",
				"gtk-4.0": ".config/gtk-4.0",
			})
		case "autostart":
			progress(pct, "Saving autostart entries...")
			err = backupDir(filepath.Join(home, ".config/autostart"), filepath.Join(bdir, "files/autostart"))
		case "terminal":
			progress(pct, "Saving terminal profiles (dconf)...")
			if !catSet["dconf"] {
				err = backupDconf(bdir)
			}
		case "monitors":
			progress(pct, "Saving monitor config...")
			err = backupSingleFile(filepath.Join(home, ".config/monitors.xml"), filepath.Join(bdir, "files/single"))
		case "nautilus":
			progress(pct, "Saving Nautilus settings...")
			err = backupDir(filepath.Join(home, ".local/share/nautilus/scripts"), filepath.Join(bdir, "files/nautilus-scripts"))
		case "mimeapps":
			progress(pct, "Saving default apps...")
			err = backupSingleFile(filepath.Join(home, ".config/mimeapps.list"), filepath.Join(bdir, "files/single"))
		case "goa":
			progress(pct, "Saving online accounts...")
			err = backupDir(filepath.Join(home, ".config/goa-1.0"), filepath.Join(bdir, "files/goa-1.0"))
		case "keyrings":
			progress(pct, "Saving keyrings...")
			err = backupDir(filepath.Join(home, ".local/share/keyrings"), filepath.Join(bdir, "files/keyrings"))
		case "desktop_files":
			progress(pct, "Saving desktop files...")
			err = backupDir(filepath.Join(home, ".local/share/applications"), filepath.Join(bdir, "files/desktop-files"))
		case "xdg_dirs":
			progress(pct, "Saving XDG user dirs...")
			for _, f := range []string{"user-dirs.dirs", "user-dirs.locale"} {
				if e := backupSingleFile(filepath.Join(home, ".config", f), filepath.Join(bdir, "files/single")); e != nil && err == nil {
					err = e
				}
			}
		}

		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", cat, err))
		}
	}

	if cfg.Description != "" {
		_ = os.WriteFile(filepath.Join(bdir, "metadata", "description.txt"), []byte(cfg.Description), 0644)
	}

	var totalSize int64
	var fileCount int
	_ = filepath.Walk(bdir, func(_ string, fi os.FileInfo, walkErr error) error {
		if walkErr != nil || fi == nil {
			return nil
		}
		if !fi.IsDir() {
			totalSize += fi.Size()
			fileCount++
		}
		return nil
	})
	result.Size = fsutil.FormatSize(totalSize)
	result.FileCount = fileCount

	// Cache stats for fast listing
	WriteStatsCache(bdir, totalSize, fileCount)

	progress(100, "Backup complete!")
	return result, nil
}

func backupMetadata(bdir, home string) error {
	metaDir := filepath.Join(bdir, "metadata")
	if err := os.MkdirAll(metaDir, 0755); err != nil {
		return err
	}

	write := func(name, content string) {
		_ = os.WriteFile(filepath.Join(metaDir, name), []byte(content), 0644)
	}

	write("timestamp.txt", time.Now().Format(time.RFC3339))
	hostname, _ := os.Hostname()
	write("hostname.txt", hostname)
	username := os.Getenv("USER")
	if username == "" {
		username = "unknown"
	}
	write("username.txt", username)
	write("exporter_version.txt", "2.0.0")

	if out, err := exec.Command("gnome-shell", "--version").Output(); err == nil {
		write("gnome_version.txt", strings.TrimSpace(string(out)))
	} else {
		write("gnome_version.txt", "unknown")
	}

	write("session_type.txt", os.Getenv("XDG_SESSION_TYPE"))

	if out, err := exec.Command("uname", "-a").Output(); err == nil {
		write("system_info.txt", strings.TrimSpace(string(out)))
	}

	if _, err := exec.LookPath("gnome-extensions"); err == nil {
		if out, err := exec.Command("gnome-extensions", "list", "--user", "--details").Output(); err == nil {
			write("extensions_list_user.txt", string(out))
		}
		if out, err := exec.Command("gnome-extensions", "list", "--enabled").Output(); err == nil {
			write("extensions_enabled.txt", string(out))
		}
		if out, err := exec.Command("gnome-extensions", "list").Output(); err == nil {
			write("extensions_list_all.txt", string(out))
		}
	}

	info := distro.Detect()
	var pkgCmd *exec.Cmd
	switch info.PackageManager {
	case "pacman":
		pkgCmd = exec.Command("bash", "-c", "pacman -Qqe 2>/dev/null | grep -iE 'gnome|gtk|mutter|gdm|nautilus|adwaita|glib'")
	case "apt":
		pkgCmd = exec.Command("bash", "-c", "dpkg -l 2>/dev/null | grep -iE 'gnome|gtk|mutter|gdm|nautilus|adwaita|glib'")
	case "dnf", "zypper":
		pkgCmd = exec.Command("bash", "-c", "rpm -qa 2>/dev/null | grep -iE 'gnome|gtk|mutter|gdm|nautilus|adwaita|glib'")
	}
	if pkgCmd != nil {
		if out, err := pkgCmd.Output(); err == nil {
			write("installed_packages.txt", string(out))
		}
	}

	return nil
}

func backupDconf(bdir string) error {
	dconfDir := filepath.Join(bdir, "dconf")
	if err := os.MkdirAll(dconfDir, 0755); err != nil {
		return err
	}

	full, err := dconf.DumpFull()
	if err != nil {
		return fmt.Errorf("dconf full dump: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dconfDir, "full_dump.ini"), []byte(full), 0644); err != nil {
		return err
	}

	for key, path := range dconf.Sections {
		section, err := dconf.DumpSection(path)
		if err != nil {
			continue
		}
		if strings.TrimSpace(section) != "" {
			_ = os.WriteFile(filepath.Join(dconfDir, key+".ini"), []byte(section), 0644)
		}
	}

	home, _ := os.UserHomeDir()
	dconfBin := filepath.Join(home, ".config/dconf/user")
	if _, err := os.Stat(dconfBin); err == nil {
		_ = fsutil.CopyFile(dconfBin, dconfDir)
	}

	return nil
}

func backupDir(src, dst string) error {
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return nil
	}
	return fsutil.CopyDir(src, dst)
}

func backupSingleFile(src, dstDir string) error {
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return nil
	}
	return fsutil.CopyFile(src, dstDir)
}

func backupMultipleDirs(home, bdir string, dirs map[string]string) error {
	var lastErr error
	for key, rel := range dirs {
		src := filepath.Join(home, rel)
		dst := filepath.Join(bdir, "files", key)
		if err := backupDir(src, dst); err != nil {
			lastErr = err
		}
	}
	return lastErr
}
