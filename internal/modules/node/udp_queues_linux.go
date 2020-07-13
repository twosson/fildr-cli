// +build !noudp_queues

package node

import (
	"errors"
	"fildr-cli/internal/log"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/procfs"
	"os"
)

type (
	udpQueuesCollector struct {
		fs     procfs.FS
		desc   *prometheus.Desc
		logger log.Logger
	}
)

func init() {
	registerCollector("udp_queues", NewUDPqueuesCollector)
}

// NewUDPqueuesCollector returns a new Collector exposing network udp queued bytes.
func NewUDPqueuesCollector(logger log.Logger) (Collector, error) {
	fs, err := procfs.NewFS(procPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open procfs: %w", err)
	}
	return &udpQueuesCollector{
		fs: fs,
		desc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "udp", "queues"),
			"Number of allocated memory in the kernel for UDP datagrams in bytes.",
			[]string{"queue", "ip"}, nil,
		),
		logger: logger,
	}, nil
}

func (c *udpQueuesCollector) Update(ch chan<- prometheus.Metric) error {

	s4, errIPv4 := c.fs.NetUDPSummary()
	if errIPv4 == nil {
		ch <- prometheus.MustNewConstMetric(c.desc, prometheus.GaugeValue, float64(s4.TxQueueLength), "tx", "v4")
		ch <- prometheus.MustNewConstMetric(c.desc, prometheus.GaugeValue, float64(s4.RxQueueLength), "rx", "v4")
	} else {
		if errors.Is(errIPv4, os.ErrNotExist) {
			c.logger.Debugf("msg", "not collecting ipv4 based metrics")
		} else {
			return fmt.Errorf("couldn't get udp queued bytes: %w", errIPv4)
		}
	}

	s6, errIPv6 := c.fs.NetUDP6Summary()
	if errIPv6 == nil {
		ch <- prometheus.MustNewConstMetric(c.desc, prometheus.GaugeValue, float64(s6.TxQueueLength), "tx", "v6")
		ch <- prometheus.MustNewConstMetric(c.desc, prometheus.GaugeValue, float64(s6.RxQueueLength), "rx", "v6")
	} else {
		if errors.Is(errIPv6, os.ErrNotExist) {
			c.logger.Debugf("msg", "not collecting ipv6 based metrics")
		} else {
			return fmt.Errorf("couldn't get udp6 queued bytes: %w", errIPv6)
		}
	}

	if errors.Is(errIPv4, os.ErrNotExist) && errors.Is(errIPv6, os.ErrNotExist) {
		return ErrNoData
	}
	return nil
}
