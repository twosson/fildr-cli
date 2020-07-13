// +build !nostat

package node

import (
	"fildr-cli/internal/log"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/procfs"
)

type statCollector struct {
	fs           procfs.FS
	intr         *prometheus.Desc
	ctxt         *prometheus.Desc
	forks        *prometheus.Desc
	btime        *prometheus.Desc
	procsRunning *prometheus.Desc
	procsBlocked *prometheus.Desc
	logger       log.Logger
}

func init() {
	registerCollector("stat", NewStatCollector)
}

// NewStatCollector returns a new Collector exposing kernel/system statistics.
func NewStatCollector(logger log.Logger) (Collector, error) {
	fs, err := procfs.NewFS(procPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open procfs: %w", err)
	}
	return &statCollector{
		fs: fs,
		intr: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "intr_total"),
			"Total number of interrupts serviced.",
			nil, nil,
		),
		ctxt: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "context_switches_total"),
			"Total number of context switches.",
			nil, nil,
		),
		forks: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "forks_total"),
			"Total number of forks.",
			nil, nil,
		),
		btime: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "boot_time_seconds"),
			"Node boot time, in unixtime.",
			nil, nil,
		),
		procsRunning: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "procs_running"),
			"Number of processes in runnable state.",
			nil, nil,
		),
		procsBlocked: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "procs_blocked"),
			"Number of processes blocked waiting for I/O to complete.",
			nil, nil,
		),
		logger: logger,
	}, nil
}

// Update implements Collector and exposes kernel and system statistics.
func (c *statCollector) Update(ch chan<- prometheus.Metric) error {
	stats, err := c.fs.Stat()
	if err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(c.intr, prometheus.CounterValue, float64(stats.IRQTotal))
	ch <- prometheus.MustNewConstMetric(c.ctxt, prometheus.CounterValue, float64(stats.ContextSwitches))
	ch <- prometheus.MustNewConstMetric(c.forks, prometheus.CounterValue, float64(stats.ProcessCreated))

	ch <- prometheus.MustNewConstMetric(c.btime, prometheus.GaugeValue, float64(stats.BootTime))

	ch <- prometheus.MustNewConstMetric(c.procsRunning, prometheus.GaugeValue, float64(stats.ProcessesRunning))
	ch <- prometheus.MustNewConstMetric(c.procsBlocked, prometheus.GaugeValue, float64(stats.ProcessesBlocked))

	return nil
}
