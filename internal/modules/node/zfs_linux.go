// +build linux
// +build !nozfs

package node

import (
	"bufio"
	"errors"
	"fildr-cli/internal/log"
	"fildr-cli/internal/pusher"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var errZFSNotAvailable = errors.New("ZFS / ZFS statistics are not available")

type zfsSysctl string

func init() {
	registerCollector("zfs", NewZFSCollector)
}

type zfsCollector struct {
	linuxProcpathBase    string
	linuxZpoolIoPath     string
	linuxZpoolObjsetPath string
	linuxPathMap         map[string]string
	logger               log.Logger
}

// NewZFSCollector returns a new Collector exposing ZFS statistics.
func NewZFSCollector(logger log.Logger) (pusher.Collector, error) {
	return &zfsCollector{
		linuxProcpathBase:    "spl/kstat/zfs",
		linuxZpoolIoPath:     "/*/io",
		linuxZpoolObjsetPath: "/*/objset-*",
		linuxPathMap: map[string]string{
			"zfs_abd":         "abdstats",
			"zfs_arc":         "arcstats",
			"zfs_dbuf":        "dbuf_stats",
			"zfs_dmu_tx":      "dmu_tx",
			"zfs_dnode":       "dnodestats",
			"zfs_fm":          "fm",
			"zfs_vdev_cache":  "vdev_cache_stats", // vdev_cache is deprecated
			"zfs_vdev_mirror": "vdev_mirror_stats",
			"zfs_xuio":        "xuio_stats", // no known consumers of the XUIO interface on Linux exist
			"zfs_zfetch":      "zfetchstats",
			"zfs_zil":         "zil",
		},
		logger: logger,
	}, nil
}

func (c *zfsCollector) Update(ch chan<- prometheus.Metric) error {
	for subsystem := range c.linuxPathMap {
		if err := c.updateZfsStats(subsystem, ch); err != nil {
			if err == errZFSNotAvailable {
				c.logger.Debugf("err", err)
				// ZFS /proc files are added as new features to ZFS arrive, it is ok to continue
				continue
			}
			return err
		}
	}

	// Pool stats
	return c.updatePoolStats(ch)
}

func (s zfsSysctl) metricName() string {
	parts := strings.Split(string(s), ".")
	return strings.Replace(parts[len(parts)-1], "-", "_", -1)
}

func (c *zfsCollector) constSysctlMetric(subsystem string, sysctl zfsSysctl, value uint64) prometheus.Metric {
	metricName := sysctl.metricName()

	return prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, metricName),
			string(sysctl),
			nil,
			nil,
		),
		prometheus.UntypedValue,
		float64(value),
	)
}

func (c *zfsCollector) constPoolMetric(poolName string, sysctl zfsSysctl, value uint64) prometheus.Metric {
	metricName := sysctl.metricName()

	return prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "zfs_zpool", metricName),
			string(sysctl),
			[]string{"zpool"},
			nil,
		),
		prometheus.UntypedValue,
		float64(value),
		poolName,
	)
}

func (c *zfsCollector) constPoolObjsetMetric(poolName string, datasetName string, sysctl zfsSysctl, value uint64) prometheus.Metric {
	metricName := sysctl.metricName()

	return prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "zfs_zpool_dataset", metricName),
			string(sysctl),
			[]string{"zpool", "dataset"},
			nil,
		),
		prometheus.UntypedValue,
		float64(value),
		poolName,
		datasetName,
	)
}

// constants from https://github.com/zfsonlinux/zfs/blob/master/lib/libspl/include/sys/kstat.h
// kept as strings for comparison thus avoiding conversion to int
const (
	// kstatDataChar   = "0"
	// kstatDataInt32  = "1"
	// kstatDataUint32 = "2"
	// kstatDataInt64  = "3"
	kstatDataUint64 = "4"
	// kstatDataLong   = "5"
	// kstatDataUlong  = "6"
	// kstatDataString = "7"
)

func (c *zfsCollector) openProcFile(path string) (*os.File, error) {
	file, err := os.Open(procFilePath(path))
	if err != nil {
		// file not found error can occur if:
		// 1. zfs module is not loaded
		// 2. zfs version does not have the feature with metrics -- ok to ignore
		c.logger.Debugf("msg", "Cannot open file for reading", "path", procFilePath(path))
		return nil, errZFSNotAvailable
	}
	return file, nil
}

func (c *zfsCollector) updateZfsStats(subsystem string, ch chan<- prometheus.Metric) error {
	file, err := c.openProcFile(filepath.Join(c.linuxProcpathBase, c.linuxPathMap[subsystem]))
	if err != nil {
		return err
	}
	defer file.Close()

	return c.parseProcfsFile(file, c.linuxPathMap[subsystem], func(s zfsSysctl, v uint64) {
		ch <- c.constSysctlMetric(subsystem, s, v)
	})
}

func (c *zfsCollector) updatePoolStats(ch chan<- prometheus.Metric) error {
	zpoolPaths, err := filepath.Glob(procFilePath(filepath.Join(c.linuxProcpathBase, c.linuxZpoolIoPath)))
	if err != nil {
		return err
	}

	if zpoolPaths == nil {
		return nil
	}

	for _, zpoolPath := range zpoolPaths {
		file, err := os.Open(zpoolPath)
		if err != nil {
			// this file should exist, but there is a race where an exporting pool can remove the files -- ok to ignore
			c.logger.Debugf("msg", "Cannot open file for reading", "path", zpoolPath)
			return errZFSNotAvailable
		}

		err = c.parsePoolProcfsFile(file, zpoolPath, func(poolName string, s zfsSysctl, v uint64) {
			ch <- c.constPoolMetric(poolName, s, v)
		})
		file.Close()
		if err != nil {
			return err
		}
	}

	zpoolObjsetPaths, err := filepath.Glob(procFilePath(filepath.Join(c.linuxProcpathBase, c.linuxZpoolObjsetPath)))
	if err != nil {
		return err
	}

	if zpoolObjsetPaths == nil {
		return nil
	}

	for _, zpoolPath := range zpoolObjsetPaths {
		file, err := os.Open(zpoolPath)
		if err != nil {
			// this file should exist, but there is a race where an exporting pool can remove the files -- ok to ignore
			c.logger.Debugf("msg", "Cannot open file for reading", "path", zpoolPath)
			return errZFSNotAvailable
		}

		err = c.parsePoolObjsetFile(file, zpoolPath, func(poolName string, datasetName string, s zfsSysctl, v uint64) {
			ch <- c.constPoolObjsetMetric(poolName, datasetName, s, v)
		})
		file.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *zfsCollector) parseProcfsFile(reader io.Reader, fmtExt string, handler func(zfsSysctl, uint64)) error {
	scanner := bufio.NewScanner(reader)

	parseLine := false
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())

		if !parseLine && len(parts) == 3 && parts[0] == "name" && parts[1] == "type" && parts[2] == "data" {
			// Start parsing from here.
			parseLine = true
			continue
		}

		if !parseLine || len(parts) < 3 {
			continue
		}

		// kstat data type (column 2) should be KSTAT_DATA_UINT64, otherwise ignore
		// TODO: when other KSTAT_DATA_* types arrive, much of this will need to be restructured
		if parts[1] == kstatDataUint64 {
			key := fmt.Sprintf("kstat.zfs.misc.%s.%s", fmtExt, parts[0])
			value, err := strconv.ParseUint(parts[2], 10, 64)
			if err != nil {
				return fmt.Errorf("could not parse expected integer value for %q", key)
			}
			handler(zfsSysctl(key), value)
		}
	}
	if !parseLine {
		return fmt.Errorf("did not parse a single %q metric", fmtExt)
	}

	return scanner.Err()
}

func (c *zfsCollector) parsePoolProcfsFile(reader io.Reader, zpoolPath string, handler func(string, zfsSysctl, uint64)) error {
	scanner := bufio.NewScanner(reader)

	parseLine := false
	var fields []string
	for scanner.Scan() {
		line := strings.Fields(scanner.Text())

		if !parseLine && len(line) >= 12 && line[0] == "nread" {
			//Start parsing from here.
			parseLine = true
			fields = make([]string, len(line))
			copy(fields, line)
			continue
		}
		if !parseLine {
			continue
		}

		zpoolPathElements := strings.Split(zpoolPath, "/")
		pathLen := len(zpoolPathElements)
		if pathLen < 2 {
			return fmt.Errorf("zpool path did not return at least two elements")
		}
		zpoolName := zpoolPathElements[pathLen-2]
		zpoolFile := zpoolPathElements[pathLen-1]

		for i, field := range fields {
			key := fmt.Sprintf("kstat.zfs.misc.%s.%s", zpoolFile, field)

			value, err := strconv.ParseUint(line[i], 10, 64)
			if err != nil {
				return fmt.Errorf("could not parse expected integer value for %q: %w", key, err)
			}
			handler(zpoolName, zfsSysctl(key), value)
		}
	}

	return scanner.Err()
}

func (c *zfsCollector) parsePoolObjsetFile(reader io.Reader, zpoolPath string, handler func(string, string, zfsSysctl, uint64)) error {
	scanner := bufio.NewScanner(reader)

	parseLine := false
	var zpoolName, datasetName string
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())

		if !parseLine && len(parts) == 3 && parts[0] == "name" && parts[1] == "type" && parts[2] == "data" {
			parseLine = true
			continue
		}

		if !parseLine || len(parts) < 3 {
			continue
		}
		if parts[0] == "dataset_name" {
			zpoolPathElements := strings.Split(zpoolPath, "/")
			pathLen := len(zpoolPathElements)
			zpoolName = zpoolPathElements[pathLen-2]
			datasetName = parts[2]
			continue
		}

		if parts[1] == kstatDataUint64 {
			key := fmt.Sprintf("kstat.zfs.misc.objset.%s", parts[0])
			value, err := strconv.ParseUint(parts[2], 10, 64)
			if err != nil {
				return fmt.Errorf("could not parse expected integer value for %q", key)
			}
			handler(zpoolName, datasetName, zfsSysctl(key), value)
		}
	}
	if !parseLine {
		return fmt.Errorf("did not parse a single %s %s metric", zpoolName, datasetName)
	}

	return scanner.Err()
}
