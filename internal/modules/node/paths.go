package node

import (
	"github.com/prometheus/procfs"
	"path/filepath"
	"strings"
)

var (
	// The path of the proc filesystem.
	procPath   = procfs.DefaultMountPoint
	sysPath    = "/sys"
	rootfsPath = "/"
)

func procFilePath(name string) string {
	return filepath.Join(procPath, name)
}

func sysFilePath(name string) string {
	return filepath.Join(sysPath, name)
}

func rootfsFilePath(name string) string {
	return filepath.Join(rootfsPath, name)
}

func rootfsStripPrefix(path string) string {
	if rootfsPath == "/" {
		return path
	}
	stripped := strings.TrimPrefix(path, rootfsPath)
	if stripped == "" {
		return "/"
	}
	return stripped
}
