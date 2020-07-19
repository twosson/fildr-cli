package lotus

import (
	"fildr-cli/internal/gateway"
	"fildr-cli/internal/log"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/prometheus/client_golang/prometheus"
	"strings"
)

type lotusDaemonCollector struct {
	version    *prometheus.Desc
	peersAddr  *prometheus.Desc
	peersCount *prometheus.Desc

	client *Client
	closer jsonrpc.ClientCloser
}

func init() {
	registerCollector("lotus-daemon", NewLotusDaemonCollector)
}

func NewLotusDaemonCollector(logger log.Logger) (gateway.Collector, error) {
	client := &Client{}
	closer, err := InitClient(client)
	if err != nil {
		return nil, err
	}

	version := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "daemon", "version"),
		"lotus daemon version.",
		[]string{"version"},
		nil,
	)

	peersAddr := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "daemon", "paddr"),
		"lotus daemon peers addr.",
		[]string{"id", "ip", "port"},
		nil,
	)

	peersCount := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "daemon", "pcount"),
		"lotus daemon peers count.",
		nil,
		nil,
	)

	return &lotusDaemonCollector{client: client, closer: closer, version: version, peersAddr: peersAddr, peersCount: peersCount}, nil
}

func (lc *lotusDaemonCollector) Update(ch chan<- prometheus.Metric) error {
	v, err := lc.client.Version()
	if err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		lc.version,
		prometheus.GaugeValue,
		1,
		v.Version,
	)

	ps, err := lc.client.NetPeers()
	if err != nil {
		return err
	}

	sc, err := lc.client.NetPubsubScores()
	if err != nil {
		return err
	}

	scm := make(map[string]float64)
	for i := range ps {
		if i < len(sc) {
			scm[sc[i].ID.String()] = sc[i].Score
		}
	}

	if len(ps) > 0 && len(sc) > 0 {
		for i := range ps {

			addrs := ps[i].Addrs[0]
			if addrs != nil {
				addrsArr := strings.Split(addrs.String(), "/")

				ch <- prometheus.MustNewConstMetric(
					lc.peersAddr,
					prometheus.GaugeValue,
					scm[ps[i].ID.String()],
					ps[i].ID.String(),
					addrsArr[2],
					addrsArr[4],
				)
			}
		}
	} else {
		ch <- prometheus.MustNewConstMetric(
			lc.peersCount,
			prometheus.GaugeValue,
			0,
		)
	}

	ch <- prometheus.MustNewConstMetric(
		lc.peersCount,
		prometheus.GaugeValue,
		float64(len(ps)),
	)

	return nil
}
