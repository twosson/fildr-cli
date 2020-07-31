package lotus

import "github.com/prometheus/client_golang/prometheus"

var (
	minerNamespace  = "miner"
	minerNumber     = "miner_number"
	ownerNumber     = "owner_number"
	minerVersion    = "version"
	minerApiVersion = "api_version"
)

var (
	minerInfoDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, minerNamespace, "info"),
		"Lotus net peers and scores",
		[]string{ownerNumber, minerNumber, minerVersion, minerApiVersion, "sector_size", "peer_id"},
		nil,
	)

	minerPeersDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, minerNamespace, "peers"),
		"Lotus net peers and scores",
		[]string{ownerNumber, minerNumber, "peer_id", "ip", "port"},
		nil,
	)

	minerConnectionsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, minerNamespace, "connections"),
		"Lotus miner net peers connections",
		[]string{ownerNumber, minerNumber},
		nil,
	)

	minerBytePowerDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, minerNamespace, "minerbytepower"),
		"Filecoin miner Byte Power",
		[]string{ownerNumber, minerNumber},
		nil,
	)

	totalBytePowerDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, minerNamespace, "totalbytepower"),
		"Filecoin total byte power",
		[]string{ownerNumber, minerNumber},
		nil,
	)

	minerActualPowerDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, minerNamespace, "mineractualpower"),
		"Filecoin miner actual power",
		[]string{ownerNumber, minerNumber},
		nil,
	)

	totalActualPowerDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, minerNamespace, "totalactualpower"),
		"Filecoin total actual power",
		[]string{ownerNumber, minerNumber},
		nil,
	)

	bytePowerRateDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, minerNamespace, "bytepowerrate"),
		"Filecoin byte power rate",
		[]string{ownerNumber, minerNumber},
		nil,
	)

	actualPowerRateDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, minerNamespace, "actualpowerrate"),
		"Filecoin actual power rate",
		[]string{ownerNumber, minerNumber},
		nil,
	)

	committedDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, minerNamespace, "committed"),
		"Filecoin actual power committed",
		[]string{ownerNumber, minerNumber},
		nil,
	)

	provingDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, minerNamespace, "proving"),
		"Filecoin actual power proving",
		[]string{ownerNumber, minerNumber},
		nil,
	)

	blockWinRateDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, minerNamespace, "blockwinrate"),
		"Filecoin block win rate",
		[]string{ownerNumber, minerNumber},
		nil,
	)

	blockWinPerDayDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, minerNamespace, "blockwinpreday"),
		"Filecoin block win per day",
		[]string{ownerNumber, minerNumber},
		nil,
	)

	minerBalanceDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, minerNamespace, "minerbalance"),
		"Filecoin miner balance",
		[]string{ownerNumber, minerNumber},
		nil,
	)

	minerBalancePreCommitDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, minerNamespace, "minerbalanceprecommit"),
		"Filecoin miner balance pre commit",
		[]string{ownerNumber, minerNumber},
		nil,
	)

	minerBalanceLockedDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, minerNamespace, "minerbalancelocked"),
		"Filecoin miner balance locked",
		[]string{ownerNumber, minerNumber},
		nil,
	)

	minerBalanceAvailableDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, minerNamespace, "minerbalanceavailable"),
		"Filecoin miner balance available",
		[]string{ownerNumber, minerNumber},
		nil,
	)

	workerBalanceDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, minerNamespace, "workerbalance"),
		"Filecoin miner worker balance",
		[]string{ownerNumber, minerNumber},
		nil,
	)

	marketEscrowBalanceDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, minerNamespace, "marketescrow"),
		"Filecoin market escrow",
		[]string{ownerNumber, minerNumber},
		nil,
	)

	marketLockedBalanceDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, minerNamespace, "marketlocked"),
		"Filecoin market escrow",
		[]string{ownerNumber, minerNumber},
		nil,
	)

	dealsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, minerNamespace, "deals"),
		"Filecoin market deals",
		[]string{ownerNumber, minerNumber, "mode"},
		nil,
	)

	faultsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, minerNamespace, "faults"),
		"Filecoin proving faults",
		[]string{ownerNumber, minerNumber},
		nil,
	)

	recoveringDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, minerNamespace, "recovering"),
		"Filecoin proving recovering",
		[]string{ownerNumber, minerNumber},
		nil,
	)

	faultPercDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, minerNamespace, "faultperc"),
		"Filecoin proving fault perc",
		[]string{ownerNumber, minerNumber},
		nil,
	)

	storageInfoDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, minerNamespace, "storageinfo"),
		"Filecoin storage info",
		[]string{ownerNumber, minerNumber, "storage_id", "weight", "seal", "store", "path", "url"},
		nil,
	)

	storageMetricsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, minerNamespace, "storagemetrics"),
		"Filecoin storage metrics",
		[]string{ownerNumber, minerNumber, "storage_id", "metric_name"},
		nil,
	)

	jobsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, minerNamespace, "jobs"),
		"Filecoin jobs",
		[]string{ownerNumber, minerNumber, "worker_id", "jobs_id", "sector_id", "hostname", "task", "start_time"},
		nil,
	)

	jobsDurationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, minerNamespace, "jobsduration"),
		"Filecoin jobs duration",
		[]string{ownerNumber, minerNumber, "task", "job_id", "worker_id"},
		nil,
	)

	workerInfoDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, minerNamespace, "workerinfo"),
		"Filecoin worker info",
		[]string{ownerNumber, minerNumber, "worker_id", "hostname", "cpus", "ram", "vmem"},
		nil,
	)

	workerMetricsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, minerNamespace, "workermetrics"),
		"Filecoin worker metrics",
		[]string{ownerNumber, minerNumber, "worker_id", "hostname", "metric_name"},
		nil,
	)

	workerGpuDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, minerNamespace, "workergpu"),
		"Filecoin worker metrics",
		[]string{ownerNumber, minerNumber, "worker_id", "hostname", "gpu_name"},
		nil,
	)
)
