// +build !nonetdev
// +build linux freebsd openbsd dragonfly darwin

package node

import (
	"bufio"
	"errors"
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

var (
	netdevDeviceInclude    = ""
	oldNetdevDeviceInclude = ""
	netdevDeviceExclude    = ""
	oldNetdevDeviceExclude = ""
)

var (
	procNetDevInterfaceRE = regexp.MustCompile(`^(.+): *(.+)$`)
	procNetDevFieldSep    = regexp.MustCompile(` +`)
)

type netDevCollector struct {
	subsystem            string
	deviceExcludePattern *regexp.Regexp
	deviceIncludePattern *regexp.Regexp
	metricDescs          map[string]*prometheus.Desc
	logger               log.Logger
}

func init() {
	registerCollector("netdev", NewNetDevCollector)
}

// NewNetDevCollector returns a new Collector exposing network device stats.
func NewNetDevCollector(logger log.Logger) (gateway.Collector, error) {
	if oldNetdevDeviceInclude != "" {
		if netdevDeviceInclude == "" {
			logger.Warnf("msg", "--collector.netdev.device-whitelist is DEPRECATED and will be removed in 2.0.0, use --collector.netdev.device-include")
			netdevDeviceInclude = oldNetdevDeviceInclude
		} else {
			return nil, errors.New("--collector.netdev.device-whitelist and --collector.netdev.device-include are mutually exclusive")
		}
	}

	if oldNetdevDeviceExclude != "" {
		if netdevDeviceExclude == "" {
			logger.Warnf("msg", "--collector.netdev.device-blacklist is DEPRECATED and will be removed in 2.0.0, use --collector.netdev.device-exclude")
			netdevDeviceExclude = oldNetdevDeviceExclude
		} else {
			return nil, errors.New("--collector.netdev.device-blacklist and --collector.netdev.device-exclude are mutually exclusive")
		}
	}

	if netdevDeviceExclude != "" && netdevDeviceInclude != "" {
		return nil, errors.New("device-exclude & device-include are mutually exclusive")
	}

	var excludePattern *regexp.Regexp
	if netdevDeviceExclude != "" {
		logger.Infof("msg", "Parsed flag --collector.netdev.device-exclude", "flag", netdevDeviceExclude)
		excludePattern = regexp.MustCompile(netdevDeviceExclude)
	}

	var includePattern *regexp.Regexp
	if netdevDeviceInclude != "" {
		logger.Infof("msg", "Parsed Flag --collector.netdev.device-include", "flag", netdevDeviceInclude)
		includePattern = regexp.MustCompile(netdevDeviceInclude)
	}

	return &netDevCollector{
		subsystem:            "network",
		deviceExcludePattern: excludePattern,
		deviceIncludePattern: includePattern,
		metricDescs:          map[string]*prometheus.Desc{},
		logger:               logger,
	}, nil
}

func (c *netDevCollector) Update(ch chan<- prometheus.Metric) error {
	netDev, err := getNetDevStats(c.deviceExcludePattern, c.deviceIncludePattern, c.logger)
	if err != nil {
		return fmt.Errorf("couldn't get netstats: %w", err)
	}
	for dev, devStats := range netDev {
		for key, value := range devStats {
			desc, ok := c.metricDescs[key]
			if !ok {
				desc = prometheus.NewDesc(
					prometheus.BuildFQName(namespace, c.subsystem, key+"_total"),
					fmt.Sprintf("Network device statistic %s.", key),
					[]string{"device"},
					nil,
				)
				c.metricDescs[key] = desc
			}
			v, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return fmt.Errorf("invalid value %s in netstats: %w", value, err)
			}
			ch <- prometheus.MustNewConstMetric(desc, prometheus.CounterValue, v, dev)
		}
	}
	return nil
}

func getNetDevStats(ignore *regexp.Regexp, accept *regexp.Regexp, logger log.Logger) (map[string]map[string]string, error) {
	file, err := os.Open(procFilePath("net/dev"))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return parseNetDevStats(file, ignore, accept, logger)
}

func parseNetDevStats(r io.Reader, ignore *regexp.Regexp, accept *regexp.Regexp, logger log.Logger) (map[string]map[string]string, error) {
	scanner := bufio.NewScanner(r)
	scanner.Scan() // skip first header
	scanner.Scan()
	parts := strings.Split(scanner.Text(), "|")
	if len(parts) != 3 { // interface + receive + transmit
		return nil, fmt.Errorf("invalid header line in net/dev: %s",
			scanner.Text())
	}

	receiveHeader := strings.Fields(parts[1])
	transmitHeader := strings.Fields(parts[2])
	headerLength := len(receiveHeader) + len(transmitHeader)

	netDev := map[string]map[string]string{}
	for scanner.Scan() {
		line := strings.TrimLeft(scanner.Text(), " ")
		parts := procNetDevInterfaceRE.FindStringSubmatch(line)
		if len(parts) != 3 {
			return nil, fmt.Errorf("couldn't get interface name, invalid line in net/dev: %q", line)
		}

		dev := parts[1]
		if ignore != nil && ignore.MatchString(dev) {
			logger.Debugf("msg", "Ignoring device", "device", dev)
			continue
		}
		if accept != nil && !accept.MatchString(dev) {
			logger.Debugf("msg", "Ignoring device", "device", dev)
			continue
		}

		values := procNetDevFieldSep.Split(strings.TrimLeft(parts[2], " "), -1)
		if len(values) != headerLength {
			return nil, fmt.Errorf("couldn't get values, invalid line in net/dev: %q", parts[2])
		}

		netDev[dev] = map[string]string{}
		for i := 0; i < len(receiveHeader); i++ {
			netDev[dev]["receive_"+receiveHeader[i]] = values[i]
		}

		for i := 0; i < len(transmitHeader); i++ {
			netDev[dev]["transmit_"+transmitHeader[i]] = values[i+len(receiveHeader)]
		}
	}
	return netDev, scanner.Err()
}
