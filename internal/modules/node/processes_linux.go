// +build !noprocesses

package node

import (
	"errors"
	"fildr-cli/internal/gateway"
	"fildr-cli/internal/log"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/procfs"
	"os"
)

type processCollector struct {
	fs          procfs.FS
	threadAlloc *prometheus.Desc
	threadLimit *prometheus.Desc
	procsState  *prometheus.Desc
	pidUsed     *prometheus.Desc
	pidMax      *prometheus.Desc
	logger      log.Logger
}

func init() {
	//registerCollector("processes", NewProcessStatCollector)
}

// NewProcessStatCollector returns a new Collector exposing process data read from the proc filesystem.
func NewProcessStatCollector(logger log.Logger) (gateway.Collector, error) {
	fs, err := procfs.NewFS(procPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open procfs: %w", err)
	}
	subsystem := "processes"
	return &processCollector{
		fs: fs,
		threadAlloc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "threads"),
			"Allocated threads in system",
			nil, nil,
		),
		threadLimit: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "max_threads"),
			"Limit of threads in the system",
			nil, nil,
		),
		procsState: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "state"),
			"Number of processes in each state.",
			[]string{"state"}, nil,
		),
		pidUsed: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "pids"),
			"Number of PIDs", nil, nil,
		),
		pidMax: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "max_processes"),
			"Number of max PIDs limit", nil, nil,
		),
		logger: logger,
	}, nil
}
func (c *processCollector) Update(ch chan<- prometheus.Metric) error {
	pids, states, threads, err := c.getAllocatedThreads()
	if err != nil {
		return fmt.Errorf("unable to retrieve number of allocated threads: %w", err)
	}

	ch <- prometheus.MustNewConstMetric(c.threadAlloc, prometheus.GaugeValue, float64(threads))
	maxThreads, err := readUintFromFile(procFilePath("sys/kernel/threads-max"))
	if err != nil {
		return fmt.Errorf("unable to retrieve limit number of threads: %w", err)
	}
	ch <- prometheus.MustNewConstMetric(c.threadLimit, prometheus.GaugeValue, float64(maxThreads))

	for state := range states {
		ch <- prometheus.MustNewConstMetric(c.procsState, prometheus.GaugeValue, float64(states[state]), state)
	}

	pidM, err := readUintFromFile(procFilePath("sys/kernel/pid_max"))
	if err != nil {
		return fmt.Errorf("unable to retrieve limit number of maximum pids alloved: %w", err)
	}
	ch <- prometheus.MustNewConstMetric(c.pidUsed, prometheus.GaugeValue, float64(pids))
	ch <- prometheus.MustNewConstMetric(c.pidMax, prometheus.GaugeValue, float64(pidM))

	return nil
}

func (c *processCollector) getAllocatedThreads() (int, map[string]int32, int, error) {
	p, err := c.fs.AllProcs()
	if err != nil {
		return 0, nil, 0, err
	}
	pids := 0
	thread := 0
	procStates := make(map[string]int32)
	for _, pid := range p {
		stat, err := pid.Stat()
		// PIDs can vanish between getting the list and getting stats.
		if errors.Is(err, os.ErrNotExist) {
			c.logger.Debugf("msg", "file not found when retrieving stats for pid", "pid", pid, "err", err)
			continue
		}
		if err != nil {
			c.logger.Debugf("msg", "error reading stat for pid", "pid", pid, "err", err)
			return 0, nil, 0, err
		}
		pids++
		procStates[stat.State]++
		thread += stat.NumThreads
	}
	return pids, procStates, thread, nil
}
