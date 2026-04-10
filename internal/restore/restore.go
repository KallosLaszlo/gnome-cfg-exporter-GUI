package restore

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gnome-cfg-exporter/internal/backup"
	"gnome-cfg-exporter/internal/dconf"
	"gnome-cfg-exporter/internal/fsutil"
)

// Config holds options for a restore run.
type Config struct {
	BackupPath string   `json:"backupPath"`
	Categories []string `json:"categories"`
	BackupDir  string   `json:"backupDir"`
}

// Result holds the outcome of a restore.
type Result struct {
	SafetyBackupPath string   `json:"safetyBackupPath"`
	RestoredCats     []string `json:"restoredCats"`
	Warnings         []string `json:"warnings"`
	Errors           []string `json:"errors"`
}

// Warnings checks for version/user mismatches.
func Warnings(backupPath string) []string {
	var warns []string
	backupVer := readMeta(backupPath, "gnome_version.txt")
	currentVer := ""
	if out, err := exec.Command("gnome-shell", "--version").Output(); err == nil {
		currentVer = strings.TrimSpace(string(out))
	}
	if backupVer != "" && currentVer != "" && backupVer != currentVer {
		warns = append(warns, fmt.Sprintf("GNOME version mismatch: backup=%s, current=%s", backupVer, currentVer))
	}
	backupUser := readMeta(backupPath, "username.txt")
	currentUser := os.Getenv("USER")
	if backupUser != "" && currentUser != "" && backupUser != currentUser {
		warns = append(warns, fmt.Sprintf("Username mismatch: backup=%s, current=%s", backupUser, currentUser))
	}
	return warns
}

// ProgressFunc is called with progress percentage and status.
type ProgressFunc func(pct int, msg string)

// Run executes a restore with the given config.
func Run(cfg Config, progress ProgressFunc) (*Result, error) {
	home, _ := os.UserHomeDir()
	result := &Result{Warnings: Warnings(cfg.BackupPath)}

	// Pre-restore safety backup
	progress(5, "Creating safety backup of current config...")
	allCats := make([]string, 0, len(backup.AllCategories()))
	for _, c := range backup.AllCategories() {
		allCats = append(allCats, c.ID)
	}
	safetyResult, err := backup.Run(backup.Config{
		Categories:  allCats,
		Description: "Pre-restore safety backup",
		BackupDir:   cfg.BackupDir,
	}, func(int, string) {})
	if err != nil {
		result.Warnings = append(result.Warnings, "Safety backup failed: "+err.Error())
	} else {
		result.SafetyBackupPath = safetyResult.Path
	}

	total := len(cfg.Categories)
	if total == 0 {
		total = 1
	}

	for i, cat := range cfg.Categories {
		pct := 10 + (i+1)*90/total
		var restoreErr error

		switch cat {
		case "dconf":
			progress(pct, "Restoring dconf database...")
			restoreErr = restoreDconf(cfg.BackupPath)
		case "extensions":
			progress(pct, "Restoring extensions...")
			restoreErr = restoreDir(filepath.Join(cfg.BackupPath, "extensions"), filepath.Join(home, ".local/share/gnome-shell/extensions"))
		case "themes":
			progress(pct, "Restoring themes...")
			restoreErr = restoreMultipleDirs(cfg.BackupPath, home, map[string]string{
				"files/themes_local":  ".local/share/themes",
				"files/themes_legacy": ".themes",
			})
		case "icons":
			progress(pct, "Restoring icons...")
			restoreErr = restoreMultipleDirs(cfg.BackupPath, home, map[string]string{
				"files/icons_local":  ".local/share/icons",
				"files/icons_legacy": ".icons",
			})
		case "fonts":
			progress(pct, "Restoring fonts...")
			restoreErr = restoreMultipleDirs(cfg.BackupPath, home, map[string]string{
				"files/fonts_local":  ".local/share/fonts",
				"files/fonts_legacy": ".fonts",
			})
			if _, err := exec.LookPath("fc-cache"); err == nil {
				_ = exec.Command("fc-cache", "-f").Run()
			}
		case "wallpapers":
			progress(pct, "Restoring wallpapers...")
			restoreErr = restoreDir(filepath.Join(cfg.BackupPath, "files/backgrounds"), filepath.Join(home, ".local/share/backgrounds"))
		case "gtk":
			progress(pct, "Restoring GTK settings...")
			restoreErr = restoreMultipleDirs(cfg.BackupPath, home, map[string]string{
				"files/gtk-3.0": ".config/gtk-3.0",
				"files/gtk-4.0": ".config/gtk-4.0",
			})
		case "autostart":
			progress(pct, "Restoring autostart entries...")
			restoreErr = restoreDir(filepath.Join(cfg.BackupPath, "files/autostart"), filepath.Join(home, ".config/autostart"))
		case "terminal":
			progress(pct, "Restoring terminal profiles (dconf)...")
			restoreErr = restoreDconfSection(cfg.BackupPath, "gnome_terminal", "/org/gnome/terminal/")
		case "monitors":
			progress(pct, "Restoring monitor config...")
			restoreErr = restoreSingleFile(filepath.Join(cfg.BackupPath, "files/single/monitors.xml"), filepath.Join(home, ".config"))
		case "nautilus":
			progress(pct, "Restoring Nautilus settings...")
			restoreErr = restoreDir(filepath.Join(cfg.BackupPath, "files/nautilus-scripts"), filepath.Join(home, ".local/share/nautilus/scripts"))
		case "mimeapps":
			progress(pct, "Restoring default apps...")
			restoreErr = restoreSingleFile(filepath.Join(cfg.BackupPath, "files/single/mimeapps.list"), filepath.Join(home, ".config"))
		case "goa":
			progress(pct, "Restoring online accounts...")
			restoreErr = restoreDir(filepath.Join(cfg.BackupPath, "files/goa-1.0"), filepath.Join(home, ".config/goa-1.0"))
			result.Warnings = append(result.Warnings, "Online Accounts may need re-authentication")
		case "keyrings":
			progress(pct, "Restoring keyrings...")
			restoreErr = restoreDir(filepath.Join(cfg.BackupPath, "files/keyrings"), filepath.Join(home, ".local/share/keyrings"))
			result.Warnings = append(result.Warnings, "Keyrings restored — some passwords may need re-entry")
		case "desktop_files":
			progress(pct, "Restoring desktop files...")
			restoreErr = restoreDir(filepath.Join(cfg.BackupPath, "files/desktop-files"), filepath.Join(home, ".local/share/applications"))
		case "xdg_dirs":
			progress(pct, "Restoring XDG user dirs...")
			for _, f := range []string{"user-dirs.dirs", "user-dirs.locale"} {
				src := filepath.Join(cfg.BackupPath, "files/single", f)
				if e := restoreSingleFile(src, filepath.Join(home, ".config")); e != nil && restoreErr == nil {
					restoreErr = e
				}
			}
		}

		if restoreErr != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", cat, restoreErr))
		} else {
			result.RestoredCats = append(result.RestoredCats, cat)
		}
	}

	progress(100, "Restore complete! Please log out and back in.")
	return result, nil
}

func restoreDconf(backupPath string) error {
	fullDump := filepath.Join(backupPath, "dconf", "full_dump.ini")
	data, err := os.ReadFile(fullDump)
	if err != nil {
		return fmt.Errorf("read dconf dump: %w", err)
	}
	return dconf.LoadFull(string(data))
}

func restoreDconfSection(backupPath, key, path string) error {
	iniFile := filepath.Join(backupPath, "dconf", key+".ini")
	data, err := os.ReadFile(iniFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return dconf.LoadSection(path, string(data))
}

func restoreDir(src, dst string) error {
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return nil
	}
	return fsutil.CopyDir(src, dst)
}

func restoreSingleFile(src, dstDir string) error {
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return nil
	}
	return fsutil.CopyFile(src, dstDir)
}

func restoreMultipleDirs(backupPath, home string, dirs map[string]string) error {
	var lastErr error
	for backupRel, homeRel := range dirs {
		src := filepath.Join(backupPath, backupRel)
		dst := filepath.Join(home, homeRel)
		if err := restoreDir(src, dst); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

func readMeta(backupPath, filename string) string {
	data, err := os.ReadFile(filepath.Join(backupPath, "metadata", filename))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}
