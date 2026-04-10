package distro

import (
	"os"
	"os/exec"
	"strings"
)

// Info holds detected distribution information.
type Info struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	PackageManager string `json:"packageManager"`
}

// Detect reads /etc/os-release and identifies the distro.
func Detect() Info {
	info := Info{ID: "unknown", Name: "Unknown Linux", PackageManager: "unknown"}
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return info
	}
	for _, line := range strings.Split(string(data), "\n") {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		val := strings.Trim(parts[1], "\"")
		switch parts[0] {
		case "ID":
			info.ID = val
		case "NAME":
			info.Name = val
		}
	}
	info.PackageManager = detectPkgManager()
	return info
}

func detectPkgManager() string {
	managers := []struct {
		cmd  string
		name string
	}{
		{"pacman", "pacman"},
		{"apt-get", "apt"},
		{"dnf", "dnf"},
		{"zypper", "zypper"},
		{"apk", "apk"},
		{"xbps-install", "xbps"},
		{"emerge", "portage"},
		{"nix-env", "nix"},
	}
	for _, m := range managers {
		if _, err := exec.LookPath(m.cmd); err == nil {
			return m.name
		}
	}
	return "unknown"
}

// MapPkgName maps a generic package name to a distro-specific one.
func MapPkgName(generic, pm string) string {
	m := map[string]map[string]string{
		"dconf": {"pacman": "dconf", "apt": "dconf-cli", "dnf": "dconf", "zypper": "dconf", "apk": "dconf", "xbps": "dconf"},
		"rsync": {"pacman": "rsync", "apt": "rsync", "dnf": "rsync", "zypper": "rsync", "apk": "rsync", "xbps": "rsync"},
		"glib":  {"pacman": "glib2", "apt": "libglib2.0-bin", "dnf": "glib2", "zypper": "glib2-tools", "apk": "glib", "xbps": "glib"},
	}
	if pkgs, ok := m[generic]; ok {
		if name, ok := pkgs[pm]; ok {
			return name
		}
	}
	return generic
}
