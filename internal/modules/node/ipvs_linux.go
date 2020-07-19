// +build !noipvs

package node

import (
	"errors"
	"fildr-cli/internal/gateway"
	"fildr-cli/internal/log"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/procfs"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"sort"
	"strconv"
	"strings"
)

type ipvsCollector struct {
	gateway.Collector
	fs                                                                          procfs.FS
	backendLabels                                                               []string
	backendConnectionsActive, backendConnectionsInact, backendWeight            gateway.TypedDesc
	connections, incomingPackets, outgoingPackets, incomingBytes, outgoingBytes gateway.TypedDesc
	logger                                                                      log.Logger
}

type ipvsBackendStatus struct {
	ActiveConn uint64
	InactConn  uint64
	Weight     uint64
}

const (
	ipvsLabelLocalAddress  = "local_address"
	ipvsLabelLocalPort     = "local_port"
	ipvsLabelRemoteAddress = "remote_address"
	ipvsLabelRemotePort    = "remote_port"
	ipvsLabelProto         = "proto"
	ipvsLabelLocalMark     = "local_mark"
)

var (
	fullIpvsBackendLabels = []string{
		ipvsLabelLocalAddress,
		ipvsLabelLocalPort,
		ipvsLabelRemoteAddress,
		ipvsLabelRemotePort,
		ipvsLabelProto,
		ipvsLabelLocalMark,
	}
	ipvsLabels = kingpin.Flag("collector.ipvs.backend-labels", "Comma separated list for IPVS backend stats labels.").Default(strings.Join(fullIpvsBackendLabels, ",")).String()
)

func init() {
	registerCollector("ipvs", NewIPVSCollector)
}

// NewIPVSCollector sets up a new collector for IPVS metrics. It accepts the
// "procfs" config parameter to override the default proc location (/proc).
func NewIPVSCollector(logger log.Logger) (gateway.Collector, error) {
	return newIPVSCollector(logger)
}

func newIPVSCollector(logger log.Logger) (*ipvsCollector, error) {
	var (
		c         ipvsCollector
		err       error
		subsystem = "ipvs"
	)

	if c.backendLabels, err = c.parseIpvsLabels(*ipvsLabels); err != nil {
		return nil, err
	}

	c.logger = logger
	c.fs, err = procfs.NewFS(procPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open procfs: %w", err)
	}

	c.connections = gateway.TypedDesc{prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, "connections_total"),
		"The total number of connections made.",
		nil, nil,
	), prometheus.CounterValue}
	c.incomingPackets = gateway.TypedDesc{prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, "incoming_packets_total"),
		"The total number of incoming packets.",
		nil, nil,
	), prometheus.CounterValue}
	c.outgoingPackets = gateway.TypedDesc{prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, "outgoing_packets_total"),
		"The total number of outgoing packets.",
		nil, nil,
	), prometheus.CounterValue}
	c.incomingBytes = gateway.TypedDesc{prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, "incoming_bytes_total"),
		"The total amount of incoming data.",
		nil, nil,
	), prometheus.CounterValue}
	c.outgoingBytes = gateway.TypedDesc{prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, "outgoing_bytes_total"),
		"The total amount of outgoing data.",
		nil, nil,
	), prometheus.CounterValue}
	c.backendConnectionsActive = gateway.TypedDesc{prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, "backend_connections_active"),
		"The current active connections by local and remote address.",
		c.backendLabels, nil,
	), prometheus.GaugeValue}
	c.backendConnectionsInact = gateway.TypedDesc{prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, "backend_connections_inactive"),
		"The current inactive connections by local and remote address.",
		c.backendLabels, nil,
	), prometheus.GaugeValue}
	c.backendWeight = gateway.TypedDesc{prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, "backend_weight"),
		"The current backend weight by local and remote address.",
		c.backendLabels, nil,
	), prometheus.GaugeValue}

	return &c, nil
}

func (c *ipvsCollector) Update(ch chan<- prometheus.Metric) error {
	ipvsStats, err := c.fs.IPVSStats()
	if err != nil {
		// Cannot access ipvs metrics, report no error.
		if errors.Is(err, os.ErrNotExist) {
			c.logger.Debugf("msg", "ipvs collector metrics are not available for this system")
			return gateway.ErrNoData
		}
		return fmt.Errorf("could not get IPVS stats: %w", err)
	}
	ch <- c.connections.MustNewConstMetric(float64(ipvsStats.Connections))
	ch <- c.incomingPackets.MustNewConstMetric(float64(ipvsStats.IncomingPackets))
	ch <- c.outgoingPackets.MustNewConstMetric(float64(ipvsStats.OutgoingPackets))
	ch <- c.incomingBytes.MustNewConstMetric(float64(ipvsStats.IncomingBytes))
	ch <- c.outgoingBytes.MustNewConstMetric(float64(ipvsStats.OutgoingBytes))

	backendStats, err := c.fs.IPVSBackendStatus()
	if err != nil {
		return fmt.Errorf("could not get backend status: %w", err)
	}

	sums := map[string]ipvsBackendStatus{}
	labelValues := map[string][]string{}
	for _, backend := range backendStats {
		localAddress := ""
		if backend.LocalAddress.String() != "<nil>" {
			localAddress = backend.LocalAddress.String()
		}
		kv := make([]string, len(c.backendLabels))
		for i, label := range c.backendLabels {
			var labelValue string
			switch label {
			case ipvsLabelLocalAddress:
				labelValue = localAddress
			case ipvsLabelLocalPort:
				labelValue = strconv.FormatUint(uint64(backend.LocalPort), 10)
			case ipvsLabelRemoteAddress:
				labelValue = backend.RemoteAddress.String()
			case ipvsLabelRemotePort:
				labelValue = strconv.FormatUint(uint64(backend.RemotePort), 10)
			case ipvsLabelProto:
				labelValue = backend.Proto
			case ipvsLabelLocalMark:
				labelValue = backend.LocalMark
			}
			kv[i] = labelValue
		}
		key := strings.Join(kv, "-")
		status := sums[key]
		status.ActiveConn += backend.ActiveConn
		status.InactConn += backend.InactConn
		status.Weight += backend.Weight
		sums[key] = status
		labelValues[key] = kv
	}
	for key, status := range sums {
		kv := labelValues[key]
		ch <- c.backendConnectionsActive.MustNewConstMetric(float64(status.ActiveConn), kv...)
		ch <- c.backendConnectionsInact.MustNewConstMetric(float64(status.InactConn), kv...)
		ch <- c.backendWeight.MustNewConstMetric(float64(status.Weight), kv...)
	}
	return nil
}

func (c *ipvsCollector) parseIpvsLabels(labelString string) ([]string, error) {
	labels := strings.Split(labelString, ",")
	labelSet := make(map[string]bool, len(labels))
	results := make([]string, 0, len(labels))
	for _, label := range labels {
		if label != "" {
			labelSet[label] = true
		}
	}

	for _, label := range fullIpvsBackendLabels {
		if labelSet[label] {
			results = append(results, label)
		}
		delete(labelSet, label)
	}

	if len(labelSet) > 0 {
		keys := make([]string, 0, len(labelSet))
		for label := range labelSet {
			keys = append(keys, label)
		}
		sort.Strings(keys)
		return nil, fmt.Errorf("unknown IPVS backend labels: %q", strings.Join(keys, ", "))
	}

	return results, nil
}
