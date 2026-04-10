package main

import (
	"embed"
	"fmt"
	"os"
	"strings"

	"gnome-cfg-exporter/internal/backup"
	"gnome-cfg-exporter/internal/config"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	wailslinux "github.com/wailsapp/wails/v2/pkg/options/linux"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--backup", "-b":
			runCLIBackup()
			return
		case "--list", "-l":
			runCLIList()
			return
		case "--help", "-h":
			printHelp()
			return
		case "--version", "-v":
			fmt.Println("GNOME Config Exporter v2.0.0")
			return
		}
	}

	app := NewApp()

	err := wails.Run(&options.App{
		Title:     "GNOME Config Exporter",
		Width:     960,
		Height:    620,
		MinWidth:  750,
		MinHeight: 480,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 30, G: 30, B: 30, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		Linux: &wailslinux.Options{
			ProgramName: "GNOME Config Exporter",
		},
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}
}

func runCLIBackup() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}
	allCats := make([]string, 0, len(backup.AllCategories()))
	for _, c := range backup.AllCategories() {
		allCats = append(allCats, c.ID)
	}
	result, err := backup.Run(backup.Config{
		Categories:  allCats,
		Description: "CLI backup",
		BackupDir:   cfg.BackupDir,
	}, func(pct int, msg string) {
		fmt.Printf("[%3d%%] %s\n", pct, msg)
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Backup failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("\nBackup complete!\n  Path:  %s\n  Size:  %s\n  Files: %d\n", result.Path, result.Size, result.FileCount)
	if len(result.Errors) > 0 {
		fmt.Printf("  Warnings: %s\n", strings.Join(result.Errors, "; "))
	}
}

func runCLIList() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}
	backups, err := backup.ListBackups(cfg.BackupDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing backups: %v\n", err)
		os.Exit(1)
	}
	if len(backups) == 0 {
		fmt.Println("No backups found.")
		return
	}
	fmt.Printf("%-22s %-20s %-10s %-8s %s\n", "DATE", "GNOME", "SIZE", "FILES", "DESCRIPTION")
	fmt.Println(strings.Repeat("-", 80))
	for _, b := range backups {
		fmt.Printf("%-22s %-20s %-10s %-8d %s\n", b.Timestamp, b.GnomeVersion, b.Size, b.FileCount, b.Description)
	}
}

func printHelp() {
	fmt.Print(`GNOME Config Exporter v2.0.0

Usage:
  gnome-cfg-exporter              Launch GUI
  gnome-cfg-exporter --backup     Full backup (CLI mode)
  gnome-cfg-exporter --list       List all backups
  gnome-cfg-exporter --help       Show this help
  gnome-cfg-exporter --version    Show version

Environment:
  GNOME_CFG_EXPORTER_DIR   Override default backup location
`)
}
