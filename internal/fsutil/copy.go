package fsutil

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// CopyDir copies a directory recursively.
func CopyDir(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("source not found: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("source is not a directory: %s", src)
	}
	if hasRsync() {
		return rsyncDir(src, dst)
	}
	return copyDirNative(src, dst)
}

// CopyFile copies a single file preserving permissions.
func CopyFile(src, dstDir string) error {
	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("source not found: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("source is a directory: %s", src)
	}
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return err
	}
	dstPath := filepath.Join(dstDir, filepath.Base(src))
	return copyFileNative(src, dstPath, info.Mode())
}

// DirSize returns the size of a directory in bytes.
func DirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

// FormatSize formats bytes into human-readable format.
func FormatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func hasRsync() bool {
	_, err := exec.LookPath("rsync")
	return err == nil
}

func rsyncDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}
	if src[len(src)-1] != '/' {
		src += "/"
	}
	if dst[len(dst)-1] != '/' {
		dst += "/"
	}
	cmd := exec.Command("rsync", "-a", "--quiet", src, dst)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("rsync failed: %w: %s", err, string(out))
	}
	return nil
}

func copyDirNative(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		if info.Mode()&os.ModeSymlink != 0 {
			link, err := os.Readlink(path)
			if err != nil {
				return nil
			}
			return os.Symlink(link, target)
		}
		return copyFileNative(path, target, info.Mode())
	})
}

func copyFileNative(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
