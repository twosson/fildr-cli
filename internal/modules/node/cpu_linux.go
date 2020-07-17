// +build !nocpu

package node

import (
	"fildr-cli/internal/log"
	"fildr-cli/internal/pusher"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/procfs"
	"path/filepath"
	"strconv"
	"sync"
)

const cpuCollectorSubsystem = "cpu"

func init() {
	registerCollector("cpu", NewCpuCollector)
}

var (
	nodeCPUSecondsDesc = prometheus.NewDesc(
		prometheus.BuildFQName("node", cpuCollectorSubsystem, "usage"),
		"Usage of the cpus in 100ms.",
		[]string{"intervel"}, nil,
	)
)

type cpuCollector struct {
	fs                 procfs.FS
	cpu                *prometheus.Desc
	cpuInfo            *prometheus.Desc
	cpuGuest           *prometheus.Desc
	cpuCoreThrottle    *prometheus.Desc
	cpuPackageThrottle *prometheus.Desc
	cpuStats           []procfs.CPUStat
	cpuStatsMutex      sync.Mutex
	logger             log.Logger
}

func NewCpuCollector(logger log.Logger) (pusher.Collector, error) {

	fs, err := procfs.NewFS(procPath)
	if err != nil {
		return nil, fmt.Errorf("failed to pen procfs: %w", err)
	}

	return &cpuCollector{
		fs: fs,
		cpu: prometheus.NewDesc(
			prometheus.BuildFQName("node", cpuCollectorSubsystem, "seconds_total"),
			"Seconds the cpus spent in each mode.",
			[]string{"cpu", "mode"}, nil,
		),
		cpuInfo: prometheus.NewDesc(
			prometheus.BuildFQName("node", cpuCollectorSubsystem, "info"),
			"CPU information from /proc/cpuinfo.",
			[]string{"package", "core", "cpu", "vendor", "family", "model", "model_name", "microcode", "stepping", "cachesize"},
			nil,
		),
		cpuGuest: prometheus.NewDesc(
			prometheus.BuildFQName("node", cpuCollectorSubsystem, "guest_seconds_total"),
			"Seconds the cpus spent in guests (VMs) for each mode.",
			[]string{"cpu", "mode"}, nil,
		),
		cpuCoreThrottle: prometheus.NewDesc(
			prometheus.BuildFQName("node", cpuCollectorSubsystem, "core_throttles_total"),
			"Number of times this cpu core has been throttled.",
			[]string{"package", "core"}, nil,
		),
		cpuPackageThrottle: prometheus.NewDesc(
			prometheus.BuildFQName("node", cpuCollectorSubsystem, "package_throttles_total"),
			"Number of times this cpu package has been throttled.",
			[]string{"package"}, nil,
		),
		logger: log.NopLogger(),
	}, nil
}

func (c *cpuCollector) Update(ch chan<- prometheus.Metric) error {
	if err := c.updateInfo(ch); err != nil {
		return err
	}
	if err := c.updateStat(ch); err != nil {
		return err
	}
	if err := c.updateThermalThrottle(ch); err != nil {
		return err
	}
	return nil
}

func (c *cpuCollector) updateInfo(ch chan<- prometheus.Metric) error {
	info, err := c.fs.CPUInfo()
	if err != nil {
		return err
	}
	for _, cpu := range info {
		ch <- prometheus.MustNewConstMetric(c.cpuInfo,
			prometheus.GaugeValue,
			1,
			cpu.PhysicalID,
			cpu.CoreID,
			strconv.Itoa(int(cpu.Processor)),
			cpu.VendorID,
			cpu.CPUFamily,
			cpu.Model,
			cpu.ModelName,
			cpu.Microcode,
			cpu.Stepping,
			cpu.CacheSize)
	}
	return nil
}

// updateThermalThrottle reads /sys/devices/system/cpu/cpu* and expose thermal throttle statistics.
func (c *cpuCollector) updateThermalThrottle(ch chan<- prometheus.Metric) error {
	cpus, err := filepath.Glob(filepath.Join("devices/system/cpu/cpu[0-9]*"))
	if err != nil {
		return err
	}

	packageThrottles := make(map[uint64]uint64)
	packageCoreThrottles := make(map[uint64]map[uint64]uint64)

	// cpu loop
	for _, cpu := range cpus {
		// See
		// https://www.kernel.org/doc/Documentation/x86/topology.txt
		// https://www.kernel.org/doc/Documentation/cputopology.txt
		// https://www.kernel.org/doc/Documentation/ABI/testing/sysfs-devices-system-cpu
		var err error
		var physicalPackageID, coreID uint64

		// topology/physical_package_id
		if physicalPackageID, err = readUintFromFile(filepath.Join(cpu, "topology", "physical_package_id")); err != nil {
			c.logger.Debugf("msg", "CPU is missing physical_package_id", "cpu", cpu)
			continue
		}
		// topology/core_id
		if coreID, err = readUintFromFile(filepath.Join(cpu, "topology", "core_id")); err != nil {
			c.logger.Debugf("msg", "CPU is missing core_id", "cpu", cpu)
			continue
		}

		// metric node_cpu_core_throttles_total
		//
		// We process this metric before the package throttles as there
		// are cpu+kernel combinations that only present core throttles
		// but no package throttles.
		// Seen e.g. on an Intel Xeon E5472 system with RHEL 6.9 kernel.
		if _, present := packageCoreThrottles[physicalPackageID]; !present {
			packageCoreThrottles[physicalPackageID] = make(map[uint64]uint64)
		}
		if _, present := packageCoreThrottles[physicalPackageID][coreID]; !present {
			// Read thermal_throttle/core_throttle_count only once
			if coreThrottleCount, err := readUintFromFile(filepath.Join(cpu, "thermal_throttle", "core_throttle_count")); err == nil {
				packageCoreThrottles[physicalPackageID][coreID] = coreThrottleCount
			} else {
				c.logger.Debugf("msg", "CPU is missing core_throttle_count", "cpu", cpu)
			}
		}

		// metric node_cpu_package_throttles_total
		if _, present := packageThrottles[physicalPackageID]; !present {
			// Read thermal_throttle/package_throttle_count only once
			if packageThrottleCount, err := readUintFromFile(filepath.Join(cpu, "thermal_throttle", "package_throttle_count")); err == nil {
				packageThrottles[physicalPackageID] = packageThrottleCount
			} else {
				c.logger.Debugf("msg", "CPU is missing package_throttle_count", "cpu", cpu)
			}
		}
	}

	for physicalPackageID, packageThrottleCount := range packageThrottles {
		ch <- prometheus.MustNewConstMetric(c.cpuPackageThrottle,
			prometheus.CounterValue,
			float64(packageThrottleCount),
			strconv.FormatUint(physicalPackageID, 10))
	}

	for physicalPackageID, coreMap := range packageCoreThrottles {
		for coreID, coreThrottleCount := range coreMap {
			ch <- prometheus.MustNewConstMetric(c.cpuCoreThrottle,
				prometheus.CounterValue,
				float64(coreThrottleCount),
				strconv.FormatUint(physicalPackageID, 10),
				strconv.FormatUint(coreID, 10))
		}
	}
	return nil
}

func (c *cpuCollector) updateStat(ch chan<- prometheus.Metric) error {
	stats, err := c.fs.Stat()
	if err != nil {
		return err
	}

	c.updateCPUStats(stats.CPU)

	// Acquire a lock to read the stats.
	c.cpuStatsMutex.Lock()
	defer c.cpuStatsMutex.Unlock()
	for cpuID, cpuStat := range c.cpuStats {
		cpuNum := strconv.Itoa(cpuID)
		ch <- prometheus.MustNewConstMetric(c.cpu, prometheus.CounterValue, cpuStat.User, cpuNum, "user")
		ch <- prometheus.MustNewConstMetric(c.cpu, prometheus.CounterValue, cpuStat.Nice, cpuNum, "nice")
		ch <- prometheus.MustNewConstMetric(c.cpu, prometheus.CounterValue, cpuStat.System, cpuNum, "system")
		ch <- prometheus.MustNewConstMetric(c.cpu, prometheus.CounterValue, cpuStat.Idle, cpuNum, "idle")
		ch <- prometheus.MustNewConstMetric(c.cpu, prometheus.CounterValue, cpuStat.Iowait, cpuNum, "iowait")
		ch <- prometheus.MustNewConstMetric(c.cpu, prometheus.CounterValue, cpuStat.IRQ, cpuNum, "irq")
		ch <- prometheus.MustNewConstMetric(c.cpu, prometheus.CounterValue, cpuStat.SoftIRQ, cpuNum, "softirq")
		ch <- prometheus.MustNewConstMetric(c.cpu, prometheus.CounterValue, cpuStat.Steal, cpuNum, "steal")

		// Guest CPU is also accounted for in cpuStat.User and cpuStat.Nice, expose these as separate metrics.
		ch <- prometheus.MustNewConstMetric(c.cpuGuest, prometheus.CounterValue, cpuStat.Guest, cpuNum, "user")
		ch <- prometheus.MustNewConstMetric(c.cpuGuest, prometheus.CounterValue, cpuStat.GuestNice, cpuNum, "nice")
	}

	return nil
}

func (c *cpuCollector) updateCPUStats(newStats []procfs.CPUStat) {
	// Acquire a lock to update the stats.
	c.cpuStatsMutex.Lock()
	defer c.cpuStatsMutex.Unlock()

	// Reset the cache if the list of CPUs has changed.
	if len(c.cpuStats) != len(newStats) {
		c.cpuStats = make([]procfs.CPUStat, len(newStats))
	}

	for i, n := range newStats {
		// If idle jumps backwards, assume we had a hotplug event and reset the stats for this CPU.
		if n.Idle < c.cpuStats[i].Idle {
			c.logger.Warnf("msg", "CPU Idle counter jumped backwards, possible hotplug event, resetting CPU stats", "cpu", i, "old_value", c.cpuStats[i].Idle, "new_value", n.Idle)
			c.cpuStats[i] = procfs.CPUStat{}
		}
		c.cpuStats[i].Idle = n.Idle

		if n.User >= c.cpuStats[i].User {
			c.cpuStats[i].User = n.User
		} else {
			c.logger.Warnf("msg", "CPU User counter jumped backwards", "cpu", i, "old_value", c.cpuStats[i].User, "new_value", n.User)
		}

		if n.Nice >= c.cpuStats[i].Nice {
			c.cpuStats[i].Nice = n.Nice
		} else {
			c.logger.Warnf("msg", "CPU Nice counter jumped backwards", "cpu", i, "old_value", c.cpuStats[i].Nice, "new_value", n.Nice)
		}

		if n.System >= c.cpuStats[i].System {
			c.cpuStats[i].System = n.System
		} else {
			c.logger.Warnf("msg", "CPU System counter jumped backwards", "cpu", i, "old_value", c.cpuStats[i].System, "new_value", n.System)
		}

		if n.Iowait >= c.cpuStats[i].Iowait {
			c.cpuStats[i].Iowait = n.Iowait
		} else {
			c.logger.Warnf("msg", "CPU Iowait counter jumped backwards", "cpu", i, "old_value", c.cpuStats[i].Iowait, "new_value", n.Iowait)
		}

		if n.IRQ >= c.cpuStats[i].IRQ {
			c.cpuStats[i].IRQ = n.IRQ
		} else {
			c.logger.Warnf("msg", "CPU IRQ counter jumped backwards", "cpu", i, "old_value", c.cpuStats[i].IRQ, "new_value", n.IRQ)
		}

		if n.SoftIRQ >= c.cpuStats[i].SoftIRQ {
			c.cpuStats[i].SoftIRQ = n.SoftIRQ
		} else {
			c.logger.Warnf("msg", "CPU SoftIRQ counter jumped backwards", "cpu", i, "old_value", c.cpuStats[i].SoftIRQ, "new_value", n.SoftIRQ)
		}

		if n.Steal >= c.cpuStats[i].Steal {
			c.cpuStats[i].Steal = n.Steal
		} else {
			c.logger.Warnf("msg", "CPU Steal counter jumped backwards", "cpu", i, "old_value", c.cpuStats[i].Steal, "new_value", n.Steal)
		}

		if n.Guest >= c.cpuStats[i].Guest {
			c.cpuStats[i].Guest = n.Guest
		} else {
			c.logger.Warnf("msg", "CPU Guest counter jumped backwards", "cpu", i, "old_value", c.cpuStats[i].Guest, "new_value", n.Guest)
		}

		if n.GuestNice >= c.cpuStats[i].GuestNice {
			c.cpuStats[i].GuestNice = n.GuestNice
		} else {
			c.logger.Warnf("msg", "CPU GuestNice counter jumped backwards", "cpu", i, "old_value", c.cpuStats[i].GuestNice, "new_value", n.GuestNice)
		}
	}
}
