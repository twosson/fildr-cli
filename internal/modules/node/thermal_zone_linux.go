// +build !nothermalzone

package node

import (
	"fildr-cli/internal/gateway"
	"fildr-cli/internal/log"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/procfs/sysfs"
)

const coolingDevice = "cooling_device"
const thermalZone = "thermal_zone"

type thermalZoneCollector struct {
	fs                    sysfs.FS
	coolingDeviceCurState *prometheus.Desc
	coolingDeviceMaxState *prometheus.Desc
	zoneTemp              *prometheus.Desc
	logger                log.Logger
}

func init() {
	registerCollector("thermal_zone", NewThermalZoneCollector)
}

// NewThermalZoneCollector returns a new Collector exposing kernel/system statistics.
func NewThermalZoneCollector(logger log.Logger) (gateway.Collector, error) {
	fs, err := sysfs.NewFS(sysPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sysfs: %w", err)
	}

	return &thermalZoneCollector{
		fs: fs,
		zoneTemp: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, thermalZone, "temp"),
			"Zone temperature in Celsius",
			[]string{"zone", "type"}, nil,
		),
		coolingDeviceCurState: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, coolingDevice, "cur_state"),
			"Current throttle state of the cooling device",
			[]string{"name", "type"}, nil,
		),
		coolingDeviceMaxState: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, coolingDevice, "max_state"),
			"Maximum throttle state of the cooling device",
			[]string{"name", "type"}, nil,
		),
		logger: logger,
	}, nil
}

func (c *thermalZoneCollector) Update(ch chan<- prometheus.Metric) error {
	thermalZones, err := c.fs.ClassThermalZoneStats()
	if err != nil {
		return err
	}

	for _, stats := range thermalZones {
		ch <- prometheus.MustNewConstMetric(
			c.zoneTemp,
			prometheus.GaugeValue,
			float64(stats.Temp)/1000.0,
			stats.Name,
			stats.Type,
		)
	}

	coolingDevices, err := c.fs.ClassCoolingDeviceStats()
	if err != nil {
		return err
	}

	for _, stats := range coolingDevices {
		ch <- prometheus.MustNewConstMetric(
			c.coolingDeviceCurState,
			prometheus.GaugeValue,
			float64(stats.CurState),
			stats.Name,
			stats.Type,
		)

		ch <- prometheus.MustNewConstMetric(
			c.coolingDeviceMaxState,
			prometheus.GaugeValue,
			float64(stats.MaxState),
			stats.Name,
			stats.Type,
		)
	}

	return nil
}
