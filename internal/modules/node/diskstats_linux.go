// +build !nodiskstats

package node

import (
	"bufio"
	"fildr-cli/internal/gateway"
	"fildr-cli/internal/log"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	diskSectorSize    = 512
	diskstatsFilename = "diskstats"
)

const (
	diskSubsystem = "disk"
)

var (
	diskLabelNames = []string{"device"}

	readsCompletedDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, diskSubsystem, "reads_completed_total"),
		"The total number of reads completed successfully.",
		diskLabelNames, nil,
	)

	readBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, diskSubsystem, "read_bytes_total"),
		"The total number of bytes read successfully.",
		diskLabelNames, nil,
	)

	writesCompletedDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, diskSubsystem, "writes_completed_total"),
		"The total number of writes completed successfully.",
		diskLabelNames, nil,
	)

	writtenBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, diskSubsystem, "written_bytes_total"),
		"The total number of bytes written successfully.",
		diskLabelNames, nil,
	)

	ioTimeSecondsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, diskSubsystem, "io_time_seconds_total"),
		"Total seconds spent doing I/Os.",
		diskLabelNames, nil,
	)

	readTimeSecondsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, diskSubsystem, "read_time_seconds_total"),
		"The total number of seconds spent by all reads.",
		diskLabelNames,
		nil,
	)

	writeTimeSecondsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, diskSubsystem, "write_time_seconds_total"),
		"This is the total number of seconds spent by all writes.",
		diskLabelNames,
		nil,
	)
)

var (
	ignoredDevices = "^(ram|loop|fd|(h|s|v|xv)d[a-z]|nvme\\d+n\\d+p)\\d+$"
)

type typedFactorDesc struct {
	desc      *prometheus.Desc
	valueType prometheus.ValueType
	factor    float64
}

func (d *typedFactorDesc) mustNewConstMetric(value float64, labels ...string) prometheus.Metric {
	if d.factor != 0 {
		value *= d.factor
	}
	return prometheus.MustNewConstMetric(d.desc, d.valueType, value, labels...)
}

type diskstatsCollector struct {
	ignoredDevicesPattern *regexp.Regexp
	descs                 []typedFactorDesc
	logger                log.Logger
}

func init() {
	registerCollector("diskstats", NewDiskstatsCollector)
}

// NewDiskstatsCollector returns a new Collector exposing disk device stats.
// Docs from https://www.kernel.org/doc/Documentation/iostats.txt
func NewDiskstatsCollector(logger log.Logger) (gateway.Collector, error) {
	var diskLabelNames = []string{"device"}

	return &diskstatsCollector{
		ignoredDevicesPattern: regexp.MustCompile(ignoredDevices),
		descs: []typedFactorDesc{
			{
				desc: readsCompletedDesc, valueType: prometheus.CounterValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, diskSubsystem, "reads_merged_total"),
					"The total number of reads merged.",
					diskLabelNames,
					nil,
				), valueType: prometheus.CounterValue,
			},
			{
				desc: readBytesDesc, valueType: prometheus.CounterValue,
				factor: diskSectorSize,
			},
			{
				desc: readTimeSecondsDesc, valueType: prometheus.CounterValue,
				factor: .001,
			},
			{
				desc: writesCompletedDesc, valueType: prometheus.CounterValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, diskSubsystem, "writes_merged_total"),
					"The number of writes merged.",
					diskLabelNames,
					nil,
				), valueType: prometheus.CounterValue,
			},
			{
				desc: writtenBytesDesc, valueType: prometheus.CounterValue,
				factor: diskSectorSize,
			},
			{
				desc: writeTimeSecondsDesc, valueType: prometheus.CounterValue,
				factor: .001,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, diskSubsystem, "io_now"),
					"The number of I/Os currently in progress.",
					diskLabelNames,
					nil,
				), valueType: prometheus.GaugeValue,
			},
			{
				desc: ioTimeSecondsDesc, valueType: prometheus.CounterValue,
				factor: .001,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, diskSubsystem, "io_time_weighted_seconds_total"),
					"The weighted # of seconds spent doing I/Os.",
					diskLabelNames,
					nil,
				), valueType: prometheus.CounterValue,
				factor: .001,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, diskSubsystem, "discards_completed_total"),
					"The total number of discards completed successfully.",
					diskLabelNames,
					nil,
				), valueType: prometheus.CounterValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, diskSubsystem, "discards_merged_total"),
					"The total number of discards merged.",
					diskLabelNames,
					nil,
				), valueType: prometheus.CounterValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, diskSubsystem, "discarded_sectors_total"),
					"The total number of sectors discarded successfully.",
					diskLabelNames,
					nil,
				), valueType: prometheus.CounterValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, diskSubsystem, "discard_time_seconds_total"),
					"This is the total number of seconds spent by all discards.",
					diskLabelNames,
					nil,
				), valueType: prometheus.CounterValue,
				factor: .001,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, diskSubsystem, "flush_requests_total"),
					"The total number of flush requests completed successfully",
					diskLabelNames,
					nil,
				), valueType: prometheus.CounterValue,
			},
			{
				desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, diskSubsystem, "flush_requests_time_seconds_total"),
					"This is the total number of seconds spent by all flush requests.",
					diskLabelNames,
					nil,
				), valueType: prometheus.CounterValue,
				factor: .001,
			},
		},
		logger: logger,
	}, nil
}

func (c *diskstatsCollector) Update(ch chan<- prometheus.Metric) error {
	diskStats, err := getDiskStats()
	if err != nil {
		return fmt.Errorf("couldn't get diskstats: %w", err)
	}

	for dev, stats := range diskStats {
		if c.ignoredDevicesPattern.MatchString(dev) {
			c.logger.Debugf("msg", "Ignoring device", "device", dev)
			continue
		}

		for i, value := range stats {
			// ignore unrecognized additional stats
			if i >= len(c.descs) {
				break
			}
			v, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return fmt.Errorf("invalid value %s in diskstats: %w", value, err)
			}
			ch <- c.descs[i].mustNewConstMetric(v, dev)
		}
	}
	return nil
}

func getDiskStats() (map[string][]string, error) {
	file, err := os.Open(procFilePath(diskstatsFilename))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return parseDiskStats(file)
}

func parseDiskStats(r io.Reader) (map[string][]string, error) {
	var (
		diskStats = map[string][]string{}
		scanner   = bufio.NewScanner(r)
	)

	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) < 4 { // we strip major, minor and dev
			return nil, fmt.Errorf("invalid line in %s: %s", procFilePath(diskstatsFilename), scanner.Text())
		}
		dev := parts[2]
		diskStats[dev] = parts[3:]
	}

	return diskStats, scanner.Err()
}
