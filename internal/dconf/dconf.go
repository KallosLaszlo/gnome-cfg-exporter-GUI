package dconf

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Sections maps section keys to their dconf paths.
var Sections = map[string]string{
	"gnome_desktop":         "/org/gnome/desktop/",
	"gnome_shell":           "/org/gnome/shell/",
	"gnome_terminal":        "/org/gnome/terminal/",
	"gnome_settings_daemon": "/org/gnome/settings-daemon/",
	"gnome_mutter":          "/org/gnome/mutter/",
	"gnome_nautilus":        "/org/gnome/nautilus/",
	"gtk_settings":          "/org/gtk/",
	"gnome_text_editor":     "/org/gnome/TextEditor/",
	"gnome_calculator":      "/org/gnome/calculator/",
	"gnome_control_center":  "/org/gnome/control-center/",
	"gnome_clocks":          "/org/gnome/clocks/",
	"gnome_weather":         "/org/gnome/Weather/",
}

// DumpFull dumps the entire dconf database.
func DumpFull() (string, error) {
	return runDconf("dump", "/")
}

// DumpSection dumps a specific dconf section.
func DumpSection(path string) (string, error) {
	if !strings.HasPrefix(path, "/") || !strings.HasSuffix(path, "/") {
		return "", fmt.Errorf("dconf path must start and end with /: %s", path)
	}
	return runDconf("dump", path)
}

// LoadFull resets all dconf and loads from INI content.
func LoadFull(ini string) error {
	if err := runDconfNoOutput("reset", "-f", "/"); err != nil {
		return fmt.Errorf("dconf reset failed: %w", err)
	}
	return loadINI("/", ini)
}

// LoadSection resets a specific section and loads INI content into it.
func LoadSection(path, ini string) error {
	if !strings.HasPrefix(path, "/") || !strings.HasSuffix(path, "/") {
		return fmt.Errorf("dconf path must start and end with /: %s", path)
	}
	if err := runDconfNoOutput("reset", "-f", path); err != nil {
		return fmt.Errorf("dconf reset for %s failed: %w", path, err)
	}
	if strings.TrimSpace(ini) == "" {
		return nil
	}
	return loadINI(path, ini)
}

func loadINI(path, ini string) error {
	cmd := exec.Command("dconf", "load", path)
	cmd.Stdin = strings.NewReader(ini)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("dconf load %s: %w: %s", path, err, stderr.String())
	}
	return nil
}

func runDconf(args ...string) (string, error) {
	cmd := exec.Command("dconf", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("dconf %s: %w: %s", strings.Join(args, " "), err, stderr.String())
	}
	return stdout.String(), nil
}

func runDconfNoOutput(args ...string) error {
	cmd := exec.Command("dconf", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("dconf %s: %w: %s", strings.Join(args, " "), err, stderr.String())
	}
	return nil
}
