// +build darwin dragonfly freebsd linux netbsd openbsd solaris
// +build !noloadavg

package node

import (
	"fildr-cli/internal/log"
	"fildr-cli/internal/pusher"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"io/ioutil"
	"strconv"
	"strings"
)

type loadavgCollector struct {
	metric []pusher.TypedDesc
	logger log.Logger
}

func init() {
	registerCollector("loadavg", NewLoadavgCollector)
}

// NewLoadavgCollector returns a new Collector exposing load average stats.
func NewLoadavgCollector(logger log.Logger) (pusher.Collector, error) {
	return &loadavgCollector{
		metric: []pusher.TypedDesc{
			{prometheus.NewDesc(namespace+"_load1", "1m load average.", nil, nil), prometheus.GaugeValue},
			{prometheus.NewDesc(namespace+"_load5", "5m load average.", nil, nil), prometheus.GaugeValue},
			{prometheus.NewDesc(namespace+"_load15", "15m load average.", nil, nil), prometheus.GaugeValue},
		},
		logger: logger,
	}, nil
}

func (c *loadavgCollector) Update(ch chan<- prometheus.Metric) error {
	loads, err := getLoad()
	if err != nil {
		return fmt.Errorf("couldn't get load: %w", err)
	}
	for i, load := range loads {
		c.logger.Debugf("msg", "return load", "index", i, "load", load)
		ch <- c.metric[i].MustNewConstMetric(load)
	}
	return err
}

// Read loadavg from /proc.
func getLoad() (loads []float64, err error) {
	data, err := ioutil.ReadFile(procFilePath("loadavg"))
	if err != nil {
		return nil, err
	}
	loads, err = parseLoad(string(data))
	if err != nil {
		return nil, err
	}
	return loads, nil
}

// Parse /proc loadavg and return 1m, 5m and 15m.
func parseLoad(data string) (loads []float64, err error) {
	loads = make([]float64, 3)
	parts := strings.Fields(data)
	if len(parts) < 3 {
		return nil, fmt.Errorf("unexpected content in %s", procFilePath("loadavg"))
	}
	for i, load := range parts[0:3] {
		loads[i], err = strconv.ParseFloat(load, 64)
		if err != nil {
			return nil, fmt.Errorf("could not parse load '%s': %w", load, err)
		}
	}
	return loads, nil
}
