// +build !noshedstat

package node

import (
	"errors"
	"fildr-cli/internal/log"
	"fildr-cli/internal/pusher"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/procfs"
	"os"
)

const nsPerSec = 1e9

var (
	runningSecondsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "schedstat", "running_seconds_total"),
		"Number of seconds CPU spent running a process.",
		[]string{"cpu"},
		nil,
	)

	waitingSecondsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "schedstat", "waiting_seconds_total"),
		"Number of seconds spent by processing waiting for this CPU.",
		[]string{"cpu"},
		nil,
	)

	timeslicesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "schedstat", "timeslices_total"),
		"Number of timeslices executed by CPU.",
		[]string{"cpu"},
		nil,
	)
)

// NewSchedstatCollector returns a new Collector exposing task scheduler statistics
func NewSchedstatCollector(logger log.Logger) (pusher.Collector, error) {
	fs, err := procfs.NewFS(procPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open procfs: %w", err)
	}

	return &schedstatCollector{fs, logger}, nil
}

type schedstatCollector struct {
	fs     procfs.FS
	logger log.Logger
}

func init() {
	registerCollector("schedstat", NewSchedstatCollector)
}

func (c *schedstatCollector) Update(ch chan<- prometheus.Metric) error {
	stats, err := c.fs.Schedstat()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			c.logger.Debugf("msg", "schedstat file does not exist")
			return pusher.ErrNoData
		}
		return err
	}

	for _, cpu := range stats.CPUs {
		ch <- prometheus.MustNewConstMetric(
			runningSecondsTotal,
			prometheus.CounterValue,
			float64(cpu.RunningNanoseconds)/nsPerSec,
			cpu.CPUNum,
		)

		ch <- prometheus.MustNewConstMetric(
			waitingSecondsTotal,
			prometheus.CounterValue,
			float64(cpu.WaitingNanoseconds)/nsPerSec,
			cpu.CPUNum,
		)

		ch <- prometheus.MustNewConstMetric(
			timeslicesTotal,
			prometheus.CounterValue,
			float64(cpu.RunTimeslices),
			cpu.CPUNum,
		)
	}

	return nil
}
