// +build !notime

package node

import (
	"fildr-cli/internal/gateway"
	"fildr-cli/internal/log"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type timeCollector struct {
	desc   *prometheus.Desc
	logger log.Logger
}

func init() {
	registerCollector("time", NewTimeCollector)
}

// NewTimeCollector returns a new Collector exposing the current system time in
// seconds since epoch.
func NewTimeCollector(logger log.Logger) (gateway.Collector, error) {
	return &timeCollector{
		desc: prometheus.NewDesc(
			namespace+"_time_seconds",
			"System time in seconds since epoch (1970).",
			nil, nil,
		),
		logger: logger,
	}, nil
}

func (c *timeCollector) Update(ch chan<- prometheus.Metric) error {
	now := float64(time.Now().UnixNano()) / 1e9
	c.logger.Debugf("msg Return time now %v", now)
	ch <- prometheus.MustNewConstMetric(c.desc, prometheus.GaugeValue, now)
	return nil
}
