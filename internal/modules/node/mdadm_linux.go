// +build !nomdadm

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

type mdadmCollector struct {
	logger log.Logger
}

func init() {
	registerCollector("mdadm", NewMdadmCollector)
}

// NewMdadmCollector returns a new Collector exposing raid statistics.
func NewMdadmCollector(logger log.Logger) (gateway.Collector, error) {
	return &mdadmCollector{logger}, nil
}

var (
	activeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "md", "state"),
		"Indicates the state of md-device.",
		[]string{"device"},
		prometheus.Labels{"state": "active"},
	)
	inActiveDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "md", "state"),
		"Indicates the state of md-device.",
		[]string{"device"},
		prometheus.Labels{"state": "inactive"},
	)
	recoveringDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "md", "state"),
		"Indicates the state of md-device.",
		[]string{"device"},
		prometheus.Labels{"state": "recovering"},
	)
	resyncDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "md", "state"),
		"Indicates the state of md-device.",
		[]string{"device"},
		prometheus.Labels{"state": "resync"},
	)

	disksDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "md", "disks"),
		"Number of active/failed/spare disks of device.",
		[]string{"device", "state"},
		nil,
	)

	disksTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "md", "disks_required"),
		"Total number of disks of device.",
		[]string{"device"},
		nil,
	)

	blocksTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "md", "blocks"),
		"Total number of blocks on device.",
		[]string{"device"},
		nil,
	)

	blocksSyncedDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "md", "blocks_synced"),
		"Number of blocks synced on device.",
		[]string{"device"},
		nil,
	)
)

func (c *mdadmCollector) Update(ch chan<- prometheus.Metric) error {
	fs, err := procfs.NewFS(procPath)

	if err != nil {
		return fmt.Errorf("failed to open procfs: %w", err)
	}

	mdStats, err := fs.MDStat()

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			c.logger.Debugf("msg", "Not collecting mdstat, file does not exist", "file", procPath)
			return gateway.ErrNoData
		}

		return fmt.Errorf("error parsing mdstatus: %w", err)
	}

	for _, mdStat := range mdStats {
		c.logger.Debugf("msg", "collecting metrics for device", "device", mdStat.Name)

		stateVals := make(map[string]float64)
		stateVals[mdStat.ActivityState] = 1

		ch <- prometheus.MustNewConstMetric(
			disksTotalDesc,
			prometheus.GaugeValue,
			float64(mdStat.DisksTotal),
			mdStat.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			disksDesc,
			prometheus.GaugeValue,
			float64(mdStat.DisksActive),
			mdStat.Name,
			"active",
		)
		ch <- prometheus.MustNewConstMetric(
			disksDesc,
			prometheus.GaugeValue,
			float64(mdStat.DisksFailed),
			mdStat.Name,
			"failed",
		)
		ch <- prometheus.MustNewConstMetric(
			disksDesc,
			prometheus.GaugeValue,
			float64(mdStat.DisksSpare),
			mdStat.Name,
			"spare",
		)
		ch <- prometheus.MustNewConstMetric(
			activeDesc,
			prometheus.GaugeValue,
			stateVals["active"],
			mdStat.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			inActiveDesc,
			prometheus.GaugeValue,
			stateVals["inactive"],
			mdStat.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			recoveringDesc,
			prometheus.GaugeValue,
			stateVals["recovering"],
			mdStat.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			resyncDesc,
			prometheus.GaugeValue,
			stateVals["resyncing"],
			mdStat.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			blocksTotalDesc,
			prometheus.GaugeValue,
			float64(mdStat.BlocksTotal),
			mdStat.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			blocksSyncedDesc,
			prometheus.GaugeValue,
			float64(mdStat.BlocksSynced),
			mdStat.Name,
		)
	}

	return nil
}
