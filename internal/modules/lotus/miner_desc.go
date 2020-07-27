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
)
