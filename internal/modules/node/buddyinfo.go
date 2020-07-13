// +build !nobuddyinfo
// +build !netbsd

package node

import (
	"fildr-cli/internal/log"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/procfs"
	"strconv"
)

const (
	buddyInfoSubsystem = "buddyinfo"
)

type buddyinfoCollector struct {
	fs     procfs.FS
	desc   *prometheus.Desc
	logger log.Logger
}

func init() {
	//registerCollector("buddyinfo", NewBuddyinfoCollector)
}

// NewBuddyinfoCollector returns a new Collector exposing buddyinfo stats.
func NewBuddyinfoCollector(logger log.Logger) (Collector, error) {
	desc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, buddyInfoSubsystem, "blocks"),
		"Count of free blocks according to size.",
		[]string{"node", "zone", "size"}, nil,
	)
	fs, err := procfs.NewFS(procPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open procfs: %w", err)
	}
	return &buddyinfoCollector{fs, desc, logger}, nil
}

// Update calls (*buddyinfoCollector).getBuddyInfo to get the platform specific
// buddyinfo metrics.
func (c *buddyinfoCollector) Update(ch chan<- prometheus.Metric) error {
	buddyInfo, err := c.fs.BuddyInfo()
	if err != nil {
		return fmt.Errorf("couldn't get buddyinfo: %w", err)
	}

	c.logger.Debugf("msg", "Set node_buddy", "buddyInfo", buddyInfo)
	for _, entry := range buddyInfo {
		for size, value := range entry.Sizes {
			ch <- prometheus.MustNewConstMetric(
				c.desc,
				prometheus.GaugeValue, value,
				entry.Node, entry.Zone, strconv.Itoa(size),
			)
		}
	}
	return nil
}
