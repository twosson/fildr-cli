// +build !nofilefd

package node

import (
	"bytes"
	"fildr-cli/internal/log"
	"fildr-cli/internal/pusher"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"io/ioutil"
	"os"
	"strconv"
)

const (
	fileFDStatSubsystem = "filefd"
)

type fileFDStatCollector struct {
	logger log.Logger
}

func init() {
	registerCollector(fileFDStatSubsystem, NewFileFDStatCollector)
}

// NewFileFDStatCollector returns a new Collector exposing file-nr stats.
func NewFileFDStatCollector(logger log.Logger) (pusher.Collector, error) {
	return &fileFDStatCollector{logger}, nil
}

func (c *fileFDStatCollector) Update(ch chan<- prometheus.Metric) error {
	fileFDStat, err := parseFileFDStats(procFilePath("sys/fs/file-nr"))
	if err != nil {
		return fmt.Errorf("couldn't get file-nr: %w", err)
	}
	for name, value := range fileFDStat {
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid value %s in file-nr: %w", value, err)
		}
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, fileFDStatSubsystem, name),
				fmt.Sprintf("File descriptor statistics: %s.", name),
				nil, nil,
			),
			prometheus.GaugeValue, v,
		)
	}
	return nil
}

func parseFileFDStats(filename string) (map[string]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	content, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	parts := bytes.Split(bytes.TrimSpace(content), []byte("\u0009"))
	if len(parts) < 3 {
		return nil, fmt.Errorf("unexpected number of file stats in %q", filename)
	}

	var fileFDStat = map[string]string{}
	// The file-nr proc is only 1 line with 3 values.
	fileFDStat["allocated"] = string(parts[0])
	// The second value is skipped as it will always be zero in linux 2.6.
	fileFDStat["maximum"] = string(parts[2])

	return fileFDStat, nil
}
