package distro

import "os/exec"

// DepStatus represents the status of a dependency.
type DepStatus struct {
	Name      string `json:"name"`
	Command   string `json:"command"`
	Installed bool   `json:"installed"`
	Required  bool   `json:"required"`
	Desc      string `json:"desc"`
}

// CheckDependencies checks all required and optional dependencies.
func CheckDependencies() []DepStatus {
	deps := []DepStatus{
		{Name: "dconf", Command: "dconf", Required: true, Desc: "Read/write GNOME settings database"},
		{Name: "gsettings", Command: "gsettings", Required: true, Desc: "GNOME settings queries"},
		{Name: "rsync", Command: "rsync", Required: true, Desc: "Efficient file copying"},
		{Name: "gnome-extensions", Command: "gnome-extensions", Required: false, Desc: "Extension listing & metadata"},
		{Name: "fc-cache", Command: "fc-cache", Required: false, Desc: "Font cache refresh"},
	}
	for i := range deps {
		_, err := exec.LookPath(deps[i].Command)
		deps[i].Installed = err == nil
	}
	return deps
}

// HasCriticalDeps returns true if all required dependencies are met.
func HasCriticalDeps() bool {
	for _, d := range CheckDependencies() {
		if d.Required && !d.Installed {
			return false
		}
	}
	return true
}
