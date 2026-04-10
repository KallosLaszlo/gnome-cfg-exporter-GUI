package scheduler

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ScheduleLevel represents a schedule configuration.
type ScheduleLevel struct {
	Level   string `json:"level"`
	Enabled bool   `json:"enabled"`
	Keep    int    `json:"keep"`
}

// Status represents current schedule state.
type Status struct {
	Levels []ScheduleLevel `json:"levels"`
}

var userSystemdDir string

func init() {
	home, _ := os.UserHomeDir()
	userSystemdDir = filepath.Join(home, ".config", "systemd", "user")
}

// GetStatus returns the current schedule status.
func GetStatus() Status {
	levels := []string{"hourly", "daily", "weekly", "monthly"}
	status := Status{}
	for _, lvl := range levels {
		timerFile := filepath.Join(userSystemdDir, serviceName(lvl)+".timer")
		_, err := os.Stat(timerFile)
		status.Levels = append(status.Levels, ScheduleLevel{
			Level:   lvl,
			Enabled: err == nil,
		})
	}
	return status
}

// Enable creates and starts a systemd user timer.
func Enable(level string, keep int, binaryPath string) error {
	if err := os.MkdirAll(userSystemdDir, 0755); err != nil {
		return err
	}
	svcName := serviceName(level)

	serviceContent := fmt.Sprintf(`[Unit]
Description=GNOME Config Exporter - %s backup
After=graphical-session.target

[Service]
Type=oneshot
ExecStart=%s --backup
Environment=DISPLAY=:0
Environment=DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/%%U/bus

[Install]
WantedBy=default.target
`, level, binaryPath)

	servicePath := filepath.Join(userSystemdDir, svcName+".service")
	if err := os.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
		return fmt.Errorf("write service file: %w", err)
	}

	onCalendar := levelToCalendar(level)
	timerContent := fmt.Sprintf(`[Unit]
Description=GNOME Config Exporter - %s backup timer

[Timer]
OnCalendar=%s
Persistent=true

[Install]
WantedBy=timers.target
`, level, onCalendar)

	timerPath := filepath.Join(userSystemdDir, svcName+".timer")
	if err := os.WriteFile(timerPath, []byte(timerContent), 0644); err != nil {
		return fmt.Errorf("write timer file: %w", err)
	}

	exec.Command("systemctl", "--user", "daemon-reload").Run()
	if err := exec.Command("systemctl", "--user", "enable", "--now", svcName+".timer").Run(); err != nil {
		return fmt.Errorf("enable timer: %w", err)
	}
	return nil
}

// Disable stops and removes the systemd user timer.
func Disable(level string) error {
	svcName := serviceName(level)
	exec.Command("systemctl", "--user", "stop", svcName+".timer").Run()
	exec.Command("systemctl", "--user", "disable", svcName+".timer").Run()
	os.Remove(filepath.Join(userSystemdDir, svcName+".timer"))
	os.Remove(filepath.Join(userSystemdDir, svcName+".service"))
	exec.Command("systemctl", "--user", "daemon-reload").Run()
	return nil
}

func serviceName(level string) string {
	return "gnome-cfg-exporter-" + strings.ToLower(level)
}

func levelToCalendar(level string) string {
	switch level {
	case "hourly":
		return "*-*-* *:00:00"
	case "daily":
		return "*-*-* 02:00:00"
	case "weekly":
		return "Mon *-*-* 02:00:00"
	case "monthly":
		return "*-*-01 02:00:00"
	default:
		return "*-*-* 02:00:00"
	}
}
