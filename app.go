package main

import (
	"context"
	"os"
	"os/exec"
	"strings"

	"gnome-cfg-exporter/internal/backup"
	"gnome-cfg-exporter/internal/config"
	"gnome-cfg-exporter/internal/distro"
	"gnome-cfg-exporter/internal/restore"
	"gnome-cfg-exporter/internal/scheduler"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// SystemInfo holds system information for the frontend.
type SystemInfo struct {
	GnomeVersion string      `json:"gnomeVersion"`
	SessionType  string      `json:"sessionType"`
	Hostname     string      `json:"hostname"`
	Distro       distro.Info `json:"distro"`
}

// GetCategories returns all available backup categories.
func (a *App) GetCategories() []backup.Category {
	return backup.AllCategories()
}

// GetBackups returns all existing backups.
func (a *App) GetBackups() ([]backup.Info, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	return backup.ListBackups(cfg.BackupDir)
}

// CreateBackup creates a new backup.
func (a *App) CreateBackup(categories []string, description string) (*backup.Result, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	result, err := backup.Run(backup.Config{
		Categories:  categories,
		Description: description,
		BackupDir:   cfg.BackupDir,
	}, func(pct int, msg string) {
		runtime.EventsEmit(a.ctx, "backup:progress", map[string]interface{}{
			"percent": pct,
			"message": msg,
		})
	})
	return result, err
}

// RestoreBackup restores from a backup.
func (a *App) RestoreBackup(backupPath string, categories []string) (*restore.Result, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	result, err := restore.Run(restore.Config{
		BackupPath: backupPath,
		Categories: categories,
		BackupDir:  cfg.BackupDir,
	}, func(pct int, msg string) {
		runtime.EventsEmit(a.ctx, "restore:progress", map[string]interface{}{
			"percent": pct,
			"message": msg,
		})
	})
	return result, err
}

// GetRestoreWarnings checks for version/user mismatches before restore.
func (a *App) GetRestoreWarnings(backupPath string) []string {
	return restore.Warnings(backupPath)
}

// DeleteBackups removes the specified backups.
func (a *App) DeleteBackups(paths []string) error {
	return backup.DeleteBackups(paths)
}

// CompareBackups returns a diff between two backups.
func (a *App) CompareBackups(pathA, pathB string) (string, error) {
	return backup.CompareBackups(pathA, pathB)
}

// BrowseBackup opens the backup directory in the file manager.
func (a *App) BrowseBackup(path string) error {
	return exec.Command("xdg-open", path).Start()
}

// GetSystemInfo returns current system information.
func (a *App) GetSystemInfo() SystemInfo {
	info := SystemInfo{
		Distro: distro.Detect(),
	}
	if out, err := exec.Command("gnome-shell", "--version").Output(); err == nil {
		info.GnomeVersion = strings.TrimSpace(string(out))
	}
	info.SessionType = os.Getenv("XDG_SESSION_TYPE")
	hostname, _ := os.Hostname()
	info.Hostname = hostname
	return info
}

// CheckDependencies returns the status of all dependencies.
func (a *App) CheckDependencies() []distro.DepStatus {
	return distro.CheckDependencies()
}

// GetConfig returns the current configuration.
func (a *App) GetConfig() (config.AppConfig, error) {
	return config.Load()
}

// SaveConfig saves the configuration.
func (a *App) SaveConfig(cfg config.AppConfig) error {
	return config.Save(cfg)
}

// ConfigExists returns true if a config file exists.
func (a *App) ConfigExists() bool {
	return config.Exists()
}

// GetScheduleStatus returns the current schedule status.
func (a *App) GetScheduleStatus() scheduler.Status {
	return scheduler.GetStatus()
}

// UpdateSchedule enables or disables scheduled backups.
func (a *App) UpdateSchedule(levels []scheduler.ScheduleLevel) error {
	binaryPath, _ := os.Executable()
	for _, lvl := range levels {
		if lvl.Enabled {
			if err := scheduler.Enable(lvl.Level, lvl.Keep, binaryPath); err != nil {
				return err
			}
		} else {
			if err := scheduler.Disable(lvl.Level); err != nil {
				return err
			}
		}
	}
	return nil
}

// GetFreeSpace returns free space on the backup partition.
func (a *App) GetFreeSpace() string {
	cfg, _ := config.Load()
	return backup.GetFreeSpace(cfg.BackupDir)
}
