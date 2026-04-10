package backup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"encoding/json"

	"gnome-cfg-exporter/internal/fsutil"
)

// Info holds metadata about a backup.
type Info struct {
	Path         string   `json:"path"`
	DirName      string   `json:"dirName"`
	Timestamp    string   `json:"timestamp"`
	GnomeVersion string   `json:"gnomeVersion"`
	Hostname     string   `json:"hostname"`
	Username     string   `json:"username"`
	SessionType  string   `json:"sessionType"`
	Size         string   `json:"size"`
	FileCount    int      `json:"fileCount"`
	Description  string   `json:"description"`
	Categories   []string `json:"categories"`
}

// ListBackups scans the backup directory and returns info for each backup, newest first.
func ListBackups(backupDir string) ([]Info, error) {
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	// Collect valid backup dirs first (cheap stat check only)
	type candidate struct {
		path    string
		dirName string
	}
	var candidates []candidate
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		bpath := filepath.Join(backupDir, e.Name())
		tsFile := filepath.Join(bpath, "metadata", "timestamp.txt")
		if _, err := os.Stat(tsFile); err != nil {
			continue
		}
		candidates = append(candidates, candidate{path: bpath, dirName: e.Name()})
	}

	// Scan all backups concurrently
	backups := make([]Info, len(candidates))
	var wg sync.WaitGroup
	for i, c := range candidates {
		wg.Add(1)
		go func(idx int, bpath, dirName string) {
			defer wg.Done()
			backups[idx] = scanBackup(bpath, dirName)
		}(i, c.path, c.dirName)
	}
	wg.Wait()

	sort.Slice(backups, func(i, j int) bool {
		return backups[i].DirName > backups[j].DirName
	})

	return backups, nil
}

// scanBackup reads metadata and uses cached stats when available.
func scanBackup(bpath, dirName string) Info {
	info := Info{
		Path:    bpath,
		DirName: dirName,
	}

	// Read metadata files (small reads, fast)
	info.Timestamp = readMetaFile(bpath, "timestamp.txt")
	info.GnomeVersion = readMetaFile(bpath, "gnome_version.txt")
	info.Hostname = readMetaFile(bpath, "hostname.txt")
	info.Username = readMetaFile(bpath, "username.txt")
	info.SessionType = readMetaFile(bpath, "session_type.txt")
	info.Description = readMetaFile(bpath, "description.txt")

	// Try cached stats first (instant), fall back to walk
	if size, count, ok := ReadStatsCache(bpath); ok {
		info.Size = fsutil.FormatSize(size)
		info.FileCount = count
	} else {
		var totalSize int64
		var fileCount int
		_ = filepath.Walk(bpath, func(_ string, fi os.FileInfo, err error) error {
			if err != nil || fi == nil {
				return nil
			}
			if !fi.IsDir() {
				totalSize += fi.Size()
				fileCount++
			}
			return nil
		})
		info.Size = fsutil.FormatSize(totalSize)
		info.FileCount = fileCount
		// Save cache for next time
		WriteStatsCache(bpath, totalSize, fileCount)
	}

	info.Categories = detectCategories(bpath)
	return info
}

// DeleteBackups removes the specified backup directories.
func DeleteBackups(paths []string) error {
	var errs []string
	for _, p := range paths {
		if err := os.RemoveAll(p); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", p, err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("errors: %s", strings.Join(errs, "; "))
	}
	return nil
}

// CompareBackups returns a diff of two backups' dconf full dumps.
func CompareBackups(pathA, pathB string) (string, error) {
	dumpA := filepath.Join(pathA, "dconf", "full_dump.ini")
	dumpB := filepath.Join(pathB, "dconf", "full_dump.ini")

	if _, err := os.Stat(dumpA); err != nil {
		return "", fmt.Errorf("backup A has no dconf dump")
	}
	if _, err := os.Stat(dumpB); err != nil {
		return "", fmt.Errorf("backup B has no dconf dump")
	}

	a, _ := os.ReadFile(dumpA)
	b, _ := os.ReadFile(dumpB)

	if string(a) == string(b) {
		return "Backups are identical (dconf)", nil
	}

	// Build simple section-by-section comparison
	secA := parseSections(string(a))
	secB := parseSections(string(b))

	var sb strings.Builder
	allKeys := make(map[string]bool)
	for k := range secA {
		allKeys[k] = true
	}
	for k := range secB {
		allKeys[k] = true
	}

	for k := range allKeys {
		vA, okA := secA[k]
		vB, okB := secB[k]
		if !okA {
			sb.WriteString(fmt.Sprintf("+ [%s] (only in B)\n", k))
		} else if !okB {
			sb.WriteString(fmt.Sprintf("- [%s] (only in A)\n", k))
		} else if vA != vB {
			sb.WriteString(fmt.Sprintf("~ [%s] (modified)\n", k))
		}
	}

	if sb.Len() == 0 {
		return "Backups differ but sections are the same (values changed within sections)", nil
	}

	return sb.String(), nil
}

// GetFreeSpace returns the free space on the partition containing the given directory.
func GetFreeSpace(dir string) string {
	var stat struct{}
	_ = stat // syscall.Statfs is linux-only, use df instead
	// Use df command for simplicity
	out, err := execCommand("df", "-h", "--output=avail", dir)
	if err != nil {
		return "unknown"
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) >= 2 {
		return strings.TrimSpace(lines[1])
	}
	return "unknown"
}

func readMetaFile(bpath, filename string) string {
	data, err := os.ReadFile(filepath.Join(bpath, "metadata", filename))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func detectCategories(bpath string) []string {
	var cats []string
	checks := map[string]string{
		"dconf":         "dconf/full_dump.ini",
		"extensions":    "extensions",
		"themes":        "files/themes_local",
		"icons":         "files/icons_local",
		"fonts":         "files/fonts_local",
		"wallpapers":    "files/backgrounds",
		"gtk":           "files/gtk-3.0",
		"autostart":     "files/autostart",
		"monitors":      "files/single/monitors.xml",
		"nautilus":      "files/nautilus-scripts",
		"mimeapps":      "files/single/mimeapps.list",
		"goa":           "files/goa-1.0",
		"keyrings":      "files/keyrings",
		"desktop_files": "files/desktop-files",
		"xdg_dirs":      "files/single/user-dirs.dirs",
	}
	for cat, relPath := range checks {
		if _, err := os.Stat(filepath.Join(bpath, relPath)); err == nil {
			cats = append(cats, cat)
		}
	}
	sort.Strings(cats)
	return cats
}

func parseSections(ini string) map[string]string {
	sections := make(map[string]string)
	var current string
	var sb strings.Builder

	for _, line := range strings.Split(ini, "\n") {
		if strings.HasPrefix(line, "[") && strings.Contains(line, "]") {
			if current != "" {
				sections[current] = sb.String()
			}
			current = line
			sb.Reset()
		} else {
			sb.WriteString(line)
			sb.WriteString("\n")
		}
	}
	if current != "" {
		sections[current] = sb.String()
	}
	return sections
}

func execCommand(name string, args ...string) (string, error) {
	out, err := exec.Command(name, args...).Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}


// statsCache is the on-disk cache for backup size/filecount.
type statsCache struct {
	Size      int64 `json:"size"`
	FileCount int   `json:"fileCount"`
}

// WriteStatsCache saves size and file count to a cache file in the backup dir.
func WriteStatsCache(bpath string, size int64, fileCount int) {
	data, _ := json.Marshal(statsCache{Size: size, FileCount: fileCount})
	_ = os.WriteFile(filepath.Join(bpath, "metadata", "stats.json"), data, 0644)
}

// ReadStatsCache reads cached size and file count. Returns ok=false if no cache.
func ReadStatsCache(bpath string) (size int64, fileCount int, ok bool) {
	data, err := os.ReadFile(filepath.Join(bpath, "metadata", "stats.json"))
	if err != nil {
		return 0, 0, false
	}
	var sc statsCache
	if err := json.Unmarshal(data, &sc); err != nil {
		return 0, 0, false
	}
	return sc.Size, sc.FileCount, true
}
