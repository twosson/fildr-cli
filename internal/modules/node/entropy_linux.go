// +build !noentropy

package node

import (
	"fildr-cli/internal/log"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
)

type entropyCollector struct {
	entropyAvail *prometheus.Desc
	logger       log.Logger
}

func init() {
	registerCollector("entropy", NewEntropyCollector)
}

// NewEntropyCollector returns a new Collector exposing entropy stats.
func NewEntropyCollector(logger log.Logger) (Collector, error) {
	return &entropyCollector{
		entropyAvail: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "entropy_available_bits"),
			"Bits of available entropy.",
			nil, nil,
		),
		logger: logger,
	}, nil
}

func (c *entropyCollector) Update(ch chan<- prometheus.Metric) error {
	value, err := readUintFromFile(procFilePath("sys/kernel/random/entropy_avail"))
	if err != nil {
		return fmt.Errorf("couldn't get entropy_avail: %w", err)
	}
	ch <- prometheus.MustNewConstMetric(
		c.entropyAvail, prometheus.GaugeValue, float64(value))

	return nil
}
