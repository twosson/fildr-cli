// +build !noqdisc

package node

import (
	"encoding/json"
	"fildr-cli/internal/gateway"
	"fildr-cli/internal/log"
	"github.com/ema/qdisc"
	"github.com/prometheus/client_golang/prometheus"
	"io/ioutil"
	"path/filepath"
)

type qdiscStatCollector struct {
	bytes      gateway.TypedDesc
	packets    gateway.TypedDesc
	drops      gateway.TypedDesc
	requeues   gateway.TypedDesc
	overlimits gateway.TypedDesc
	qlength    gateway.TypedDesc
	backlog    gateway.TypedDesc
	logger     log.Logger
}

var (
	collectorQdisc = ""
)

func init() {
	//registerCollector("qdisc", NewQdiscStatCollector)
}

// NewQdiscStatCollector returns a new Collector exposing queuing discipline statistics.
func NewQdiscStatCollector(logger log.Logger) (gateway.Collector, error) {
	return &qdiscStatCollector{
		bytes: gateway.TypedDesc{prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "qdisc", "bytes_total"),
			"Number of bytes sent.",
			[]string{"device", "kind"}, nil,
		), prometheus.CounterValue},
		packets: gateway.TypedDesc{prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "qdisc", "packets_total"),
			"Number of packets sent.",
			[]string{"device", "kind"}, nil,
		), prometheus.CounterValue},
		drops: gateway.TypedDesc{prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "qdisc", "drops_total"),
			"Number of packets dropped.",
			[]string{"device", "kind"}, nil,
		), prometheus.CounterValue},
		requeues: gateway.TypedDesc{prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "qdisc", "requeues_total"),
			"Number of packets dequeued, not transmitted, and requeued.",
			[]string{"device", "kind"}, nil,
		), prometheus.CounterValue},
		overlimits: gateway.TypedDesc{prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "qdisc", "overlimits_total"),
			"Number of overlimit packets.",
			[]string{"device", "kind"}, nil,
		), prometheus.CounterValue},
		qlength: gateway.TypedDesc{prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "qdisc", "current_queue_length"),
			"Number of packets currently in queue to be sent.",
			[]string{"device", "kind"}, nil,
		), prometheus.GaugeValue},
		backlog: gateway.TypedDesc{prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "qdisc", "backlog"),
			"Number of bytes currently in queue to be sent.",
			[]string{"device", "kind"}, nil,
		), prometheus.GaugeValue},
		logger: logger,
	}, nil
}

func testQdiscGet(fixtures string) ([]qdisc.QdiscInfo, error) {
	var res []qdisc.QdiscInfo

	b, err := ioutil.ReadFile(filepath.Join(fixtures, "results.json"))
	if err != nil {
		return res, err
	}

	err = json.Unmarshal(b, &res)
	return res, err
}

func (c *qdiscStatCollector) Update(ch chan<- prometheus.Metric) error {
	var msgs []qdisc.QdiscInfo
	var err error

	fixtures := collectorQdisc

	if fixtures == "" {
		msgs, err = qdisc.Get()
	} else {
		msgs, err = testQdiscGet(fixtures)
	}

	if err != nil {
		return err
	}

	for _, msg := range msgs {
		// Only report root qdisc information.
		if msg.Parent != 0 {
			continue
		}

		ch <- c.bytes.MustNewConstMetric(float64(msg.Bytes), msg.IfaceName, msg.Kind)
		ch <- c.packets.MustNewConstMetric(float64(msg.Packets), msg.IfaceName, msg.Kind)
		ch <- c.drops.MustNewConstMetric(float64(msg.Drops), msg.IfaceName, msg.Kind)
		ch <- c.requeues.MustNewConstMetric(float64(msg.Requeues), msg.IfaceName, msg.Kind)
		ch <- c.overlimits.MustNewConstMetric(float64(msg.Overlimits), msg.IfaceName, msg.Kind)
		ch <- c.qlength.MustNewConstMetric(float64(msg.Qlen), msg.IfaceName, msg.Kind)
		ch <- c.backlog.MustNewConstMetric(float64(msg.Backlog), msg.IfaceName, msg.Kind)
	}

	return nil
}
