// +build !norapl

package node

import (
	"fildr-cli/internal/gateway"
	"fildr-cli/internal/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/procfs/sysfs"
	"strconv"
)

type raplCollector struct {
	fs sysfs.FS
}

func init() {
	registerCollector("rapl", NewRaplCollector)
}

// NewRaplCollector returns a new Collector exposing RAPL metrics.
func NewRaplCollector(logger log.Logger) (gateway.Collector, error) {
	fs, err := sysfs.NewFS(sysPath)

	if err != nil {
		return nil, err
	}

	collector := raplCollector{
		fs: fs,
	}
	return &collector, nil
}

// Update implements Collector and exposes RAPL related metrics.
func (c *raplCollector) Update(ch chan<- prometheus.Metric) error {
	// nil zones are fine when platform doesn't have powercap files present.
	zones, err := sysfs.GetRaplZones(c.fs)
	if err != nil {
		return nil
	}

	for _, rz := range zones {
		newMicrojoules, err := rz.GetEnergyMicrojoules()
		if err != nil {
			return err
		}
		index := strconv.Itoa(rz.Index)

		descriptor := prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "rapl", rz.Name+"_joules_total"),
			"Current RAPL "+rz.Name+" value in joules",
			[]string{"index"}, nil,
		)

		ch <- prometheus.MustNewConstMetric(
			descriptor,
			prometheus.CounterValue,
			float64(newMicrojoules)/1000000.0,
			index,
		)
	}
	return nil
}
