package lotus

import (
	"fildr-cli/internal/gateway"
	"fildr-cli/internal/log"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
)

type daemonCollector struct {
	daemon *daemon
	logger log.Logger
}

func init() {
	registerCollector("lotus-daemon", NewDaemonCollector)
}

func NewDaemonCollector(logger log.Logger) (gateway.Collector, error) {
	daemon, err := newDaemon()
	if err != nil {
		return nil, err
	}
	return &daemonCollector{daemon: daemon, logger: logger}, nil
}

func (d *daemonCollector) Update(ch chan<- prometheus.Metric) error {

	var runState float64 = 1
	if d.daemon.isShutdown {
		runState = 0
	}

	ch <- prometheus.MustNewConstMetric(
		daemonInfoDesc,
		prometheus.GaugeValue,
		runState,
		d.daemon.id,
		d.daemon.daemonVersion,
		d.daemon.apiVersion,
	)

	peers, err := d.daemon.lotusClient.daemonClient.api.NetPeers()
	if err != nil {
		d.logger.Warnf("Lotus net peers err: %v", err)
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		connectionsDesc,
		prometheus.GaugeValue,
		float64(len(peers)),
		d.daemon.id,
	)

	scores, err := d.daemon.lotusClient.daemonClient.api.NetPubsubScores()
	if err != nil {
		d.logger.Warnf("Lotus net pubsub scores err: %v", err)
		return err
	}

	scoresMap := make(map[peer.ID]float64, len(scores))
	for _, score := range scores {
		scoresMap[score.ID] = score.Score.Score
	}

	for _, peer := range peers {

		ip, port, err := addrsToIpAndPort(peer.Addrs)
		if err != nil {
			continue
		}

		score, ok := scoresMap[peer.ID]
		if !ok {
			score = 0
		}

		ch <- prometheus.MustNewConstMetric(
			peersDesc,
			prometheus.GaugeValue,
			score,
			d.daemon.id,
			peer.ID.Pretty(),
			ip,
			port,
		)
	}

	syncState, err := d.daemon.lotusClient.daemonClient.api.SyncState()
	if err != nil {
		d.logger.Warnf("Lotus sync status err: %v", err)
		return err
	}

	for i, sync := range syncState.ActiveSyncs {
		ch <- prometheus.MustNewConstMetric(
			syncStateDesc,
			prometheus.GaugeValue,
			float64(sync.Stage),
			d.daemon.id,
			strconv.Itoa(i),
		)

		ch <- prometheus.MustNewConstMetric(
			syncHeightDesc,
			prometheus.GaugeValue,
			float64(sync.Height),
			d.daemon.id,
			strconv.Itoa(i),
		)

		ch <- prometheus.MustNewConstMetric(
			syncHeightDiffDesc,
			prometheus.GaugeValue,
			float64(sync.Target.Height()-sync.Height),
			d.daemon.id,
			strconv.Itoa(i),
		)
	}

	return nil
}
