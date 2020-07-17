// +build !nopressure

package node

import (
	"fildr-cli/internal/log"
	"fildr-cli/internal/pusher"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/procfs"
)

var (
	psiResources = []string{"cpu", "io", "memory"}
)

type pressureStatsCollector struct {
	cpu     *prometheus.Desc
	io      *prometheus.Desc
	ioFull  *prometheus.Desc
	mem     *prometheus.Desc
	memFull *prometheus.Desc

	fs procfs.FS

	logger log.Logger
}

func init() {
	registerCollector("pressure", NewPressureStatsCollector)
}

// NewPressureStatsCollector returns a Collector exposing pressure stall information
func NewPressureStatsCollector(logger log.Logger) (pusher.Collector, error) {
	fs, err := procfs.NewFS(procPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open procfs: %w", err)
	}

	return &pressureStatsCollector{
		cpu: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "pressure", "cpu_waiting_seconds_total"),
			"Total time in seconds that processes have waited for CPU time",
			nil, nil,
		),
		io: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "pressure", "io_waiting_seconds_total"),
			"Total time in seconds that processes have waited due to IO congestion",
			nil, nil,
		),
		ioFull: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "pressure", "io_stalled_seconds_total"),
			"Total time in seconds no process could make progress due to IO congestion",
			nil, nil,
		),
		mem: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "pressure", "memory_waiting_seconds_total"),
			"Total time in seconds that processes have waited for memory",
			nil, nil,
		),
		memFull: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "pressure", "memory_stalled_seconds_total"),
			"Total time in seconds no process could make progress due to memory congestion",
			nil, nil,
		),
		fs:     fs,
		logger: logger,
	}, nil
}

// Update calls procfs.NewPSIStatsForResource for the different resources and updates the values
func (c *pressureStatsCollector) Update(ch chan<- prometheus.Metric) error {
	for _, res := range psiResources {
		c.logger.Debugf("msg", "collecting statistics for resource", "resource", res)
		vals, err := c.fs.PSIStatsForResource(res)
		if err != nil {
			c.logger.Debugf("msg", "pressure information is unavailable, you need a Linux kernel >= 4.20 and/or CONFIG_PSI enabled for your kernel")
			return nil
		}
		switch res {
		case "cpu":
			ch <- prometheus.MustNewConstMetric(c.cpu, prometheus.CounterValue, float64(vals.Some.Total)/1000.0/1000.0)
		case "io":
			ch <- prometheus.MustNewConstMetric(c.io, prometheus.CounterValue, float64(vals.Some.Total)/1000.0/1000.0)
			ch <- prometheus.MustNewConstMetric(c.ioFull, prometheus.CounterValue, float64(vals.Full.Total)/1000.0/1000.0)
		case "memory":
			ch <- prometheus.MustNewConstMetric(c.mem, prometheus.CounterValue, float64(vals.Some.Total)/1000.0/1000.0)
			ch <- prometheus.MustNewConstMetric(c.memFull, prometheus.CounterValue, float64(vals.Full.Total)/1000.0/1000.0)
		default:
			c.logger.Debugf("msg", "did not account for resource", "resource", res)
		}
	}

	return nil
}
