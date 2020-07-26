package lotus

import "github.com/prometheus/client_golang/prometheus"

var (
	daemonNamespace  = "daemon"
	daemonIdentity   = "identity"
	daemonVersion    = "version"
	daemonApiVersion = "api_version"
)

var (
	daemonInfoDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, daemonNamespace, "info"),
		"Lotus net peers and scores",
		[]string{daemonIdentity, daemonVersion, daemonApiVersion},
		nil,
	)

	peersDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, daemonNamespace, "peers"),
		"Lotus net peers and scores",
		[]string{daemonIdentity, "peer_id", "ip", "port"},
		nil,
	)

	connectionsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, daemonNamespace, "connections"),
		"Lotus net peers connections",
		[]string{daemonIdentity},
		nil,
	)

	syncStateDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, daemonNamespace, "syncstate"),
		"Lotus sync state",
		[]string{daemonIdentity, "worker_id"},
		nil,
	)

	syncHeightDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, daemonNamespace, "syncheight"),
		"Lotus sync height",
		[]string{daemonIdentity, "worker_id"},
		nil,
	)

	syncHeightDiffDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, daemonNamespace, "syncdiff"),
		"Lotus sync height diff",
		[]string{daemonIdentity, "worker_id"},
		nil,
	)
)
