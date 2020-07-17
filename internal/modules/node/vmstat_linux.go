// +build !novmstat

package node

import (
	"bufio"
	"fildr-cli/internal/log"
	"fildr-cli/internal/pusher"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	vmStatSubsystem = "vmstat"
)

var (
	vmStatFields = "^(oom_kill|pgpg|pswp|pg.*fault).*"
)

type vmStatCollector struct {
	fieldPattern *regexp.Regexp
	logger       log.Logger
}

func init() {
	registerCollector("vmstat", NewvmStatCollector)
}

// NewvmStatCollector returns a new Collector exposing vmstat stats.
func NewvmStatCollector(logger log.Logger) (pusher.Collector, error) {
	pattern := regexp.MustCompile(vmStatFields)
	return &vmStatCollector{
		fieldPattern: pattern,
		logger:       logger,
	}, nil
}

func (c *vmStatCollector) Update(ch chan<- prometheus.Metric) error {
	file, err := os.Open(procFilePath("vmstat"))
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		value, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			return err
		}
		if !c.fieldPattern.MatchString(parts[0]) {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, vmStatSubsystem, parts[0]),
				fmt.Sprintf("/proc/vmstat information field %s.", parts[0]),
				nil, nil),
			prometheus.UntypedValue,
			value,
		)
	}
	return scanner.Err()
}
