// +build !nofilesystem

package node

import (
	"bufio"
	"errors"
	"fildr-cli/internal/log"
	"fmt"
	"golang.org/x/sys/unix"
	"gopkg.in/alecthomas/kingpin.v2"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	defIgnoredMountPoints = "^/(dev|proc|sys|var/lib/docker/.+)($|/)"
	defIgnoredFSTypes     = "^(autofs|binfmt_misc|bpf|cgroup2?|configfs|debugfs|devpts|devtmpfs|fusectl|hugetlbfs|iso9660|mqueue|nsfs|overlay|proc|procfs|pstore|rpc_pipefs|securityfs|selinuxfs|squashfs|sysfs|tracefs)$"
)

var mountTimeout = kingpin.Flag("collector.filesystem.mount-timeout",
	"how long to wait for a mount to respond before marking it as stale").
	Hidden().Default("5s").Duration()
var stuckMounts = make(map[string]struct{})
var stuckMountsMtx = &sync.Mutex{}

// GetStats returns filesystem stats.
func (c *filesystemCollector) GetStats() ([]filesystemStats, error) {
	mps, err := mountPointDetails(c.logger)
	if err != nil {
		return nil, err
	}
	stats := []filesystemStats{}
	for _, labels := range mps {
		if c.ignoredMountPointsPattern.MatchString(labels.mountPoint) {
			c.logger.Debugf("msg", "Ignoring mount point", "mountpoint", labels.mountPoint)
			continue
		}
		if c.ignoredFSTypesPattern.MatchString(labels.fsType) {
			c.logger.Debugf("msg", "Ignoring fs", "type", labels.fsType)
			continue
		}
		stuckMountsMtx.Lock()
		if _, ok := stuckMounts[labels.mountPoint]; ok {
			stats = append(stats, filesystemStats{
				labels:      labels,
				deviceError: 1,
			})
			c.logger.Debugf("msg", "Mount point is in an unresponsive state", "mountpoint", labels.mountPoint)
			stuckMountsMtx.Unlock()
			continue
		}
		stuckMountsMtx.Unlock()

		// The success channel is used do tell the "watcher" that the stat
		// finished successfully. The channel is closed on success.
		success := make(chan struct{})
		go stuckMountWatcher(labels.mountPoint, success, c.logger)

		buf := new(unix.Statfs_t)
		err = unix.Statfs(rootfsFilePath(labels.mountPoint), buf)
		stuckMountsMtx.Lock()
		close(success)
		// If the mount has been marked as stuck, unmark it and log it's recovery.
		if _, ok := stuckMounts[labels.mountPoint]; ok {
			c.logger.Debugf("msg", "Mount point has recovered, monitoring will resume", "mountpoint", labels.mountPoint)
			delete(stuckMounts, labels.mountPoint)
		}
		stuckMountsMtx.Unlock()

		if err != nil {
			stats = append(stats, filesystemStats{
				labels:      labels,
				deviceError: 1,
			})

			c.logger.Debugf("msg", "Error on statfs() system call", "rootfs", rootfsFilePath(labels.mountPoint), "err", err)
			continue
		}

		var ro float64
		for _, option := range strings.Split(labels.options, ",") {
			if option == "ro" {
				ro = 1
				break
			}
		}

		stats = append(stats, filesystemStats{
			labels:    labels,
			size:      float64(buf.Blocks) * float64(buf.Bsize),
			free:      float64(buf.Bfree) * float64(buf.Bsize),
			avail:     float64(buf.Bavail) * float64(buf.Bsize),
			files:     float64(buf.Files),
			filesFree: float64(buf.Ffree),
			ro:        ro,
		})
	}
	return stats, nil
}

// stuckMountWatcher listens on the given success channel and if the channel closes
// then the watcher does nothing. If instead the timeout is reached, the
// mount point that is being watched is marked as stuck.
func stuckMountWatcher(mountPoint string, success chan struct{}, logger log.Logger) {
	select {
	case <-success:
		// Success
	case <-time.After(*mountTimeout):
		// Timed out, mark mount as stuck
		stuckMountsMtx.Lock()
		select {
		case <-success:
			// Success came in just after the timeout was reached, don't label the mount as stuck
		default:
			log.NopLogger().Warnf("msg", "Mount point timed out, it is being labeled as stuck and will not be monitored", "mountpoint", mountPoint)
			stuckMounts[mountPoint] = struct{}{}
		}
		stuckMountsMtx.Unlock()
	}
}

func mountPointDetails(logger log.Logger) ([]filesystemLabels, error) {
	file, err := os.Open(procFilePath("1/mounts"))
	if errors.Is(err, os.ErrNotExist) {
		// Fallback to `/proc/mounts` if `/proc/1/mounts` is missing due hidepid.
		logger.Debugf("msg", "Reading root mounts failed, falling back to system mounts", "err", err)
		file, err = os.Open(procFilePath("mounts"))
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return parseFilesystemLabels(file)
}

func parseFilesystemLabels(r io.Reader) ([]filesystemLabels, error) {
	var filesystems []filesystemLabels

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())

		if len(parts) < 4 {
			return nil, fmt.Errorf("malformed mount point information: %q", scanner.Text())
		}

		// Ensure we handle the translation of \040 and \011
		// as per fstab(5).
		parts[1] = strings.Replace(parts[1], "\\040", " ", -1)
		parts[1] = strings.Replace(parts[1], "\\011", "\t", -1)

		filesystems = append(filesystems, filesystemLabels{
			device:     parts[0],
			mountPoint: rootfsStripPrefix(parts[1]),
			fsType:     parts[2],
			options:    parts[3],
		})
	}

	return filesystems, scanner.Err()
}
