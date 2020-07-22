package lotus

import (
	"fildr-cli/internal/config"
	"fildr-cli/internal/gateway"
	"fildr-cli/internal/log"
	"github.com/prometheus/client_golang/prometheus"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type lotusMinerCollector struct {

	// Info
	infoIsCanWonBlocks *prometheus.Desc
	infoWinRate        *prometheus.Desc
	infoSectorSize     *prometheus.Desc
	infoBytePower      *prometheus.Desc
	infoActualPower    *prometheus.Desc
	infoCommitted      *prometheus.Desc
	infoProving        *prometheus.Desc
	infoMinerBalance   *prometheus.Desc
	infoWorkerBalance  *prometheus.Desc
	infoMarket         *prometheus.Desc
	infoSectorStat     *prometheus.Desc

	// Worker
	workerCpuUse   *prometheus.Desc
	workerRam      *prometheus.Desc
	workerVmem     *prometheus.Desc
	workerGpuIsUse *prometheus.Desc

	logger log.Logger
}

func init() {
	registerCollector("lotus-miner", NewLotusMinerCollector)
}

func NewLotusMinerCollector(logger log.Logger) (gateway.Collector, error) {

	lmc := &lotusMinerCollector{logger: logger}

	lmc.infoIsCanWonBlocks = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "miner", "canwonblocks"),
		"lotus miner info is can won blocks.",
		[]string{"miner", "daemonVersion", "minerVersion"},
		nil,
	)

	lmc.infoWinRate = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "miner", "winrate"),
		"lotus miner info win rate.",
		[]string{"miner", "daemonVersion", "minerVersion"},
		nil,
	)

	lmc.infoSectorSize = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "miner", "sectorsize"),
		"lotus miner info sector size.",
		[]string{"miner", "daemonVersion", "minerVersion"},
		nil,
	)

	lmc.infoBytePower = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "miner", "bytepower"),
		"lotus miner info byte power.",
		[]string{"miner", "actor", "daemonVersion", "minerVersion"},
		nil,
	)

	lmc.infoActualPower = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "miner", "actualpower"),
		"lotus miner info actual power.",
		[]string{"miner", "daemonVersion", "minerVersion", "actor"},
		nil,
	)

	lmc.infoCommitted = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "miner", "committed"),
		"lotus miner info actual power committed.",
		[]string{"miner", "daemonVersion", "minerVersion"},
		nil,
	)

	lmc.infoProving = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "miner", "proving"),
		"lotus miner info actual power proving.",
		[]string{"miner", "daemonVersion", "minerVersion"},
		nil,
	)

	lmc.infoMinerBalance = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "miner", "minerbalance"),
		"lotus miner info miner balance.",
		[]string{"miner", "daemonVersion", "minerVersion", "actor"},
		nil,
	)

	lmc.infoWorkerBalance = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "miner", "workerbalance"),
		"lotus miner info worker balance.",
		[]string{"miner", "daemonVersion", "minerVersion"},
		nil,
	)

	lmc.infoMarket = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "miner", "market"),
		"lotus miner info market.",
		[]string{"miner", "daemonVersion", "minerVersion", "actor"},
		nil,
	)

	lmc.infoSectorStat = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "miner", "sectorstat"),
		"lotus miner info sector stat.",
		[]string{"miner", "daemonVersion", "minerVersion", "actor", "color"},
		nil,
	)

	lmc.workerCpuUse = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "miner", "workercpu"),
		"lotus miner workers list cpu",
		[]string{"miner", "daemonVersion", "minerVersion", "hostname", "workerId"},
		nil,
	)

	lmc.workerRam = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "miner", "workerram"),
		"lotus miner workers list ram",
		[]string{"miner", "daemonVersion", "minerVersion", "hostname", "workerId", "actor"},
		nil,
	)

	lmc.workerVmem = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "miner", "workervmem"),
		"lotus miner workers list vmem",
		[]string{"miner", "daemonVersion", "minerVersion", "hostname", "workerId", "actor"},
		nil,
	)

	lmc.workerGpuIsUse = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "miner", "workergpu"),
		"lotus miner workers list vmem",
		[]string{"miner", "daemonVersion", "minerVersion", "hostname", "workerId", "gpuModel"},
		nil,
	)

	return lmc, nil
}

func (lc *lotusMinerCollector) Update(ch chan<- prometheus.Metric) error {

	cfg := config.Get()

	if !cfg.Lotus.Miner.Enable {
		return nil
	}

	out, err := exec.Command(cfg.Lotus.Miner.Path, "version").Output()

	if err != nil {
		lc.logger.Errorf("Please confirm that the miner path is correct")
		return err
	}

	outStr := string(out)
	if len(outStr) < 10 {
		return gateway.ErrNoData
	}

	var daemonVersion string
	var minerVersion string

	lines := strings.Split(outStr, "\n")
	for _, line := range lines {
		row := strings.Split(line, ":")
		colNum := len(row)
		if colNum == 2 {
			label := row[0]
			label = strings.TrimSpace(label)
			value := row[1]
			value = strings.TrimSpace(value)
			switch label {
			case "Daemon":
				daemonVersion = value
			case "Local":
				subs := strings.Split(value, " ")
				if len(subs) == 3 {
					minerVersion = subs[2]
				}
			}
		}
	}

	out, err = exec.Command(cfg.Lotus.Miner.Path, "info").Output()
	if err != nil {
		lc.logger.Errorf("Please confirm that the miner path is correct")
		return err
	}
	outStr = string(out)
	if len(outStr) < 10 {
		return gateway.ErrNoData
	}
	lines = strings.Split(outStr, "\n")

	var minerNumber string

	for _, line := range lines {
		row := strings.Split(line, ":")
		colNum := len(row)
		if colNum == 1 {
			if strings.EqualFold(line, "Below minimum power threshold, no blocks will be won") {
				ch <- prometheus.MustNewConstMetric(
					lc.infoIsCanWonBlocks,
					prometheus.GaugeValue,
					0,
					minerNumber,
					daemonVersion,
					minerVersion,
				)
			}
		} else if colNum == 2 {
			label := row[0]
			label = strings.TrimSpace(label)
			value := row[1]
			value = strings.ReplaceAll(value, " ", "")

			if strings.EqualFold(label, "Proving") && strings.HasSuffix(value, "B") {
				label = "ProvingBytes"
			}

			switch label {
			case "Miner":
				minerNumber = value
			case "Expected block win rate":
				ch <- prometheus.MustNewConstMetric(
					lc.infoIsCanWonBlocks,
					prometheus.GaugeValue,
					1,
					minerNumber,
					daemonVersion,
					minerVersion,
				)

				subs := strings.Split(value, "/")
				if len(subs) > 1 {
					value = subs[0]
					result, err := strconv.ParseFloat(value, 64)
					if err != nil {
						lc.logger.Warnf("get lotus miner info win rate err: %v", err)
					}
					ch <- prometheus.MustNewConstMetric(
						lc.infoWinRate,
						prometheus.GaugeValue,
						result,
						minerNumber,
						daemonVersion,
						minerVersion,
					)
				}
			case "Sector Size":
				ch <- prometheus.MustNewConstMetric(
					lc.infoSectorSize,
					prometheus.GaugeValue,
					float64(DescSizeStr(value)),
					minerNumber,
					daemonVersion,
					minerVersion,
				)
			case "Byte Power":
				bytePowers := strings.Split(value, "/")
				if len(bytePowers) == 2 {
					ch <- prometheus.MustNewConstMetric(
						lc.infoBytePower,
						prometheus.GaugeValue,
						float64(DescSizeStr(bytePowers[0])),
						minerNumber,
						daemonVersion,
						minerVersion,
						"owner",
					)
					totalBytePowers := strings.Split(bytePowers[1], `(`)
					if len(totalBytePowers) == 2 {
						ch <- prometheus.MustNewConstMetric(
							lc.infoBytePower,
							prometheus.GaugeValue,
							float64(DescSizeStr(totalBytePowers[0])),
							minerNumber,
							daemonVersion,
							minerVersion,
							"total",
						)
						powerProportionStr := strings.ReplaceAll(totalBytePowers[1], `%)`, "")
						powerProportion, err := strconv.ParseFloat(powerProportionStr, 64)
						if err != nil {
							lc.logger.Warnf("get byte power proportion err: %v", err)
						} else {
							ch <- prometheus.MustNewConstMetric(
								lc.infoBytePower,
								prometheus.GaugeValue,
								powerProportion,
								minerNumber,
								daemonVersion,
								minerVersion,
								"proportion",
							)
						}
					}
				}
			case "Actual Power":
				actualPowers := strings.Split(value, "/")
				if len(actualPowers) == 2 {
					ch <- prometheus.MustNewConstMetric(
						lc.infoActualPower,
						prometheus.GaugeValue,
						float64(DescDeciStr(actualPowers[0])),
						minerNumber,
						daemonVersion,
						minerVersion,
						"owner",
					)
					totalActualPowers := strings.Split(actualPowers[1], `(`)
					if len(totalActualPowers) == 2 {
						ch <- prometheus.MustNewConstMetric(
							lc.infoActualPower,
							prometheus.GaugeValue,
							float64(DescDeciStr(totalActualPowers[0])),
							minerNumber,
							daemonVersion,
							minerVersion,
							"total",
						)
						powerProportionStr := strings.ReplaceAll(totalActualPowers[1], `%)`, "")
						powerProportion, err := strconv.ParseFloat(powerProportionStr, 64)
						if err != nil {
							lc.logger.Warnf("get byte power proportion err: %v", err)
						} else {
							ch <- prometheus.MustNewConstMetric(
								lc.infoActualPower,
								prometheus.GaugeValue,
								powerProportion,
								minerNumber,
								daemonVersion,
								minerVersion,
								"proportion",
							)
						}
					}
				}
			case "Committed":
				ch <- prometheus.MustNewConstMetric(
					lc.infoCommitted,
					prometheus.GaugeValue,
					float64(DescSizeStr(value)),
					minerNumber,
					daemonVersion,
					minerVersion,
				)
			case "ProvingBytes":
				ch <- prometheus.MustNewConstMetric(
					lc.infoProving,
					prometheus.GaugeValue,
					float64(DescSizeStr(value)),
					minerNumber,
					daemonVersion,
					minerVersion,
				)
			case "Miner Balance":
				balance, err := strconv.ParseFloat(value, 64)
				if err != nil {
					lc.logger.Warnf("get lotus miner info miner balance err: %v", err)
				} else {
					ch <- prometheus.MustNewConstMetric(
						lc.infoMinerBalance,
						prometheus.GaugeValue,
						balance,
						minerNumber,
						daemonVersion,
						minerVersion,
						"miner",
					)
				}
			case "PreCommit":
				balance, err := strconv.ParseFloat(value, 64)
				if err != nil {
					lc.logger.Warnf("get lotus miner info pre commit err: %v", err)
				} else {
					ch <- prometheus.MustNewConstMetric(
						lc.infoMinerBalance,
						prometheus.GaugeValue,
						balance,
						minerNumber,
						daemonVersion,
						minerVersion,
						"commit",
					)
				}
			case "Locked":
				balance, err := strconv.ParseFloat(value, 64)
				if err != nil {
					lc.logger.Warnf("get lotus miner info locked balance err: %v", err)
				} else {
					ch <- prometheus.MustNewConstMetric(
						lc.infoMinerBalance,
						prometheus.GaugeValue,
						balance,
						minerNumber,
						daemonVersion,
						minerVersion,
						"locked",
					)
				}
			case "Available":
				balance, err := strconv.ParseFloat(value, 64)
				if err != nil {
					lc.logger.Warnf("get lotus miner info available balance err: %v", err)
				} else {
					ch <- prometheus.MustNewConstMetric(
						lc.infoMinerBalance,
						prometheus.GaugeValue,
						balance,
						minerNumber,
						daemonVersion,
						minerVersion,
						"available",
					)
				}
			case "Worker Balance":
				balance, err := strconv.ParseFloat(value, 64)
				if err != nil {
					lc.logger.Warnf("get lotus miner info worker balance err: %v", err)
				} else {
					ch <- prometheus.MustNewConstMetric(
						lc.infoWorkerBalance,
						prometheus.GaugeValue,
						balance,
						minerNumber,
						daemonVersion,
						minerVersion,
					)
				}
			case "Market (Escrow)":
				result, err := strconv.ParseFloat(value, 64)
				if err != nil {
					lc.logger.Warnf("get lotus miner info market escrow err: %v", err)
				} else {
					ch <- prometheus.MustNewConstMetric(
						lc.infoMarket,
						prometheus.GaugeValue,
						result,
						minerNumber,
						daemonVersion,
						minerVersion,
						"escrow",
					)
				}
			case "Market (Locked)":
				result, err := strconv.ParseFloat(value, 64)
				if err != nil {
					lc.logger.Warnf("get lotus miner info market locked err: %v", err)
				} else {
					ch <- prometheus.MustNewConstMetric(
						lc.infoMarket,
						prometheus.GaugeValue,
						result,
						minerNumber,
						daemonVersion,
						minerVersion,
						"locked",
					)
				}
			case "Total":
				result, err := strconv.ParseFloat(value, 64)
				if err != nil {
					lc.logger.Warnf("get lotus miner info sector total err: %v", err)
				} else {
					ch <- prometheus.MustNewConstMetric(
						lc.infoSectorStat,
						prometheus.GaugeValue,
						result,
						minerNumber,
						daemonVersion,
						minerVersion,
						"Total",
						"green",
					)
				}
			case "Proving":
				result, err := strconv.ParseFloat(value, 64)
				if err != nil {
					lc.logger.Warnf("get lotus miner info sector proving err: %v", err)
				} else {
					ch <- prometheus.MustNewConstMetric(
						lc.infoSectorStat,
						prometheus.GaugeValue,
						result,
						minerNumber,
						daemonVersion,
						minerVersion,
						"Proving",
						"green",
					)
				}
			case "UndefinedSectorState":
				result, err := strconv.ParseFloat(value, 64)
				if err != nil {
					lc.logger.Warnf("get lotus miner info sector undefined sector state err: %v", err)
				} else {
					ch <- prometheus.MustNewConstMetric(
						lc.infoSectorStat,
						prometheus.GaugeValue,
						result,
						minerNumber,
						daemonVersion,
						minerVersion,
						"UndefinedSectorState",
						"red",
					)
				}
			case "Empty":
				result, err := strconv.ParseFloat(value, 64)
				if err != nil {
					lc.logger.Warnf("get lotus miner info sector empty err: %v", err)
				} else {
					ch <- prometheus.MustNewConstMetric(
						lc.infoSectorStat,
						prometheus.GaugeValue,
						result,
						minerNumber,
						daemonVersion,
						minerVersion,
						"Empty",
						"yellow",
					)
				}
			case "Packing":
				result, err := strconv.ParseFloat(value, 64)
				if err != nil {
					lc.logger.Warnf("get lotus miner info sector packing err: %v", err)
				} else {
					ch <- prometheus.MustNewConstMetric(
						lc.infoSectorStat,
						prometheus.GaugeValue,
						result,
						minerNumber,
						daemonVersion,
						minerVersion,
						"Packing",
						"yellow",
					)
				}
			case "PreCommit1":
				result, err := strconv.ParseFloat(value, 64)
				if err != nil {
					lc.logger.Warnf("get lotus miner info sector pre commit1 err: %v", err)
				} else {
					ch <- prometheus.MustNewConstMetric(
						lc.infoSectorStat,
						prometheus.GaugeValue,
						result,
						minerNumber,
						daemonVersion,
						minerVersion,
						"PreCommit1",
						"yellow",
					)
				}
			case "PreCommit2":
				result, err := strconv.ParseFloat(value, 64)
				if err != nil {
					lc.logger.Warnf("get lotus miner info sector pre commit2 err: %v", err)
				} else {
					ch <- prometheus.MustNewConstMetric(
						lc.infoSectorStat,
						prometheus.GaugeValue,
						result,
						minerNumber,
						daemonVersion,
						minerVersion,
						"PreCommit2",
						"yellow",
					)
				}
			case "PreCommitting":
				result, err := strconv.ParseFloat(value, 64)
				if err != nil {
					lc.logger.Warnf("get lotus miner info sector pre committing err: %v", err)
				} else {
					ch <- prometheus.MustNewConstMetric(
						lc.infoSectorStat,
						prometheus.GaugeValue,
						result,
						minerNumber,
						daemonVersion,
						minerVersion,
						"PreCommitting",
						"yellow",
					)
				}
			case "PreCommitWait":
				result, err := strconv.ParseFloat(value, 64)
				if err != nil {
					lc.logger.Warnf("get lotus miner info sector pre commit wait err: %v", err)
				} else {
					ch <- prometheus.MustNewConstMetric(
						lc.infoSectorStat,
						prometheus.GaugeValue,
						result,
						minerNumber,
						daemonVersion,
						minerVersion,
						"PreCommitWait",
						"yellow",
					)
				}
			case "WaitSeed":
				result, err := strconv.ParseFloat(value, 64)
				if err != nil {
					lc.logger.Warnf("get lotus miner info sector wait seed err: %v", err)
				} else {
					ch <- prometheus.MustNewConstMetric(
						lc.infoSectorStat,
						prometheus.GaugeValue,
						result,
						minerNumber,
						daemonVersion,
						minerVersion,
						"WaitSeed",
						"yellow",
					)
				}
			case "Committing":
				result, err := strconv.ParseFloat(value, 64)
				if err != nil {
					lc.logger.Warnf("get lotus miner info sector committing err: %v", err)
				} else {
					ch <- prometheus.MustNewConstMetric(
						lc.infoSectorStat,
						prometheus.GaugeValue,
						result,
						minerNumber,
						daemonVersion,
						minerVersion,
						"Committing",
						"yellow",
					)
				}
			case "CommitWait":
				result, err := strconv.ParseFloat(value, 64)
				if err != nil {
					lc.logger.Warnf("get lotus miner info sector commit wait err: %v", err)
				} else {
					ch <- prometheus.MustNewConstMetric(
						lc.infoSectorStat,
						prometheus.GaugeValue,
						result,
						minerNumber,
						daemonVersion,
						minerVersion,
						"CommitWait",
						"yellow",
					)
				}
			case "FinalizeSector":
				result, err := strconv.ParseFloat(value, 64)
				if err != nil {
					lc.logger.Warnf("get lotus miner info sector finalize sector err: %v", err)
				} else {
					ch <- prometheus.MustNewConstMetric(
						lc.infoSectorStat,
						prometheus.GaugeValue,
						result,
						minerNumber,
						daemonVersion,
						minerVersion,
						"FinalizeSector",
						"yellow",
					)
				}
			case "FailedUnrecoverable":
				result, err := strconv.ParseFloat(value, 64)
				if err != nil {
					lc.logger.Warnf("get lotus miner info sector failed unrecoverable err: %v", err)
				} else {
					ch <- prometheus.MustNewConstMetric(
						lc.infoSectorStat,
						prometheus.GaugeValue,
						result,
						minerNumber,
						daemonVersion,
						minerVersion,
						"FailedUnrecoverable",
						"red",
					)
				}
			case "SealPreCommit1Failed":
				result, err := strconv.ParseFloat(value, 64)
				if err != nil {
					lc.logger.Warnf("get lotus miner info sector seal pre commit1 failed err: %v", err)
				} else {
					ch <- prometheus.MustNewConstMetric(
						lc.infoSectorStat,
						prometheus.GaugeValue,
						result,
						minerNumber,
						daemonVersion,
						minerVersion,
						"SealPreCommit1Failed",
						"red",
					)
				}
			case "SealPreCommit2Failed":
				result, err := strconv.ParseFloat(value, 64)
				if err != nil {
					lc.logger.Warnf("get lotus miner info sector seal pre commit2 failed err: %v", err)
				} else {
					ch <- prometheus.MustNewConstMetric(
						lc.infoSectorStat,
						prometheus.GaugeValue,
						result,
						minerNumber,
						daemonVersion,
						minerVersion,
						"SealPreCommit2Failed",
						"red",
					)
				}
			case "PreCommitFailed":
				result, err := strconv.ParseFloat(value, 64)
				if err != nil {
					lc.logger.Warnf("get lotus miner info sector pre commit failed err: %v", err)
				} else {
					ch <- prometheus.MustNewConstMetric(
						lc.infoSectorStat,
						prometheus.GaugeValue,
						result,
						minerNumber,
						daemonVersion,
						minerVersion,
						"PreCommitFailed",
						"red",
					)
				}
			case "ComputeProofFailed":
				result, err := strconv.ParseFloat(value, 64)
				if err != nil {
					lc.logger.Warnf("get lotus miner info sector compute proof failed err: %v", err)
				} else {
					ch <- prometheus.MustNewConstMetric(
						lc.infoSectorStat,
						prometheus.GaugeValue,
						result,
						minerNumber,
						daemonVersion,
						minerVersion,
						"ComputeProofFailed",
						"red",
					)
				}
			case "CommitFailed":
				result, err := strconv.ParseFloat(value, 64)
				if err != nil {
					lc.logger.Warnf("get lotus miner info sector commit failed err: %v", err)
				} else {
					ch <- prometheus.MustNewConstMetric(
						lc.infoSectorStat,
						prometheus.GaugeValue,
						result,
						minerNumber,
						daemonVersion,
						minerVersion,
						"CommitFailed",
						"red",
					)
				}
			case "PackingFailed":
				result, err := strconv.ParseFloat(value, 64)
				if err != nil {
					lc.logger.Warnf("get lotus miner info sector packing failed err: %v", err)
				} else {
					ch <- prometheus.MustNewConstMetric(
						lc.infoSectorStat,
						prometheus.GaugeValue,
						result,
						minerNumber,
						daemonVersion,
						minerVersion,
						"PackingFailed",
						"red",
					)
				}
			case "FinalizeFailed":
				result, err := strconv.ParseFloat(value, 64)
				if err != nil {
					lc.logger.Warnf("get lotus miner info sector finalize failed err: %v", err)
				} else {
					ch <- prometheus.MustNewConstMetric(
						lc.infoSectorStat,
						prometheus.GaugeValue,
						result,
						minerNumber,
						daemonVersion,
						minerVersion,
						"FinalizeFailed",
						"red",
					)
				}
			case "Faulty":
				result, err := strconv.ParseFloat(value, 64)
				if err != nil {
					lc.logger.Warnf("get lotus miner info sector faulty err: %v", err)
				} else {
					ch <- prometheus.MustNewConstMetric(
						lc.infoSectorStat,
						prometheus.GaugeValue,
						result,
						minerNumber,
						daemonVersion,
						minerVersion,
						"Faulty",
						"red",
					)
				}
			case "FaultReported":
				result, err := strconv.ParseFloat(value, 64)
				if err != nil {
					lc.logger.Warnf("get lotus miner info sector fault reported err: %v", err)
				} else {
					ch <- prometheus.MustNewConstMetric(
						lc.infoSectorStat,
						prometheus.GaugeValue,
						result,
						minerNumber,
						daemonVersion,
						minerVersion,
						"FaultReported",
						"red",
					)
				}
			case "FaultedFinal":
				result, err := strconv.ParseFloat(value, 64)
				if err != nil {
					lc.logger.Warnf("get lotus miner info sector fault final err: %v", err)
				} else {
					ch <- prometheus.MustNewConstMetric(
						lc.infoSectorStat,
						prometheus.GaugeValue,
						result,
						minerNumber,
						daemonVersion,
						minerVersion,
						"FaultedFinal",
						"red",
					)
				}
			}
		}
	}

	out, err = exec.Command(cfg.Lotus.Miner.Path, "workers", "list").Output()
	if err != nil {
		lc.logger.Errorf("Please confirm that the miner path is correct")
		return err
	}
	outStr = string(out)
	if len(outStr) < 10 {
		return gateway.ErrNoData
	}

	workers := strings.Split(outStr, "Worker ")

	for _, worker := range workers {
		if len(worker) < 10 {
			continue
		}
		lines := strings.Split(worker, "\n")
		var workerId string
		var workerHostname string
		for i, line := range lines {
			if len(line) < 5 {
				continue
			}
			if i == 0 {
				titles := strings.Split(line, ",")
				workerId = titles[0]
				workerHostname = strings.TrimSpace(strings.ReplaceAll(titles[1], "host ", ""))
			} else {
				row := strings.Split(line, ":")
				label := row[0]
				label = strings.TrimSpace(label)
				value := row[1]
				if !strings.EqualFold(label, "GPU") {
					value = strings.ReplaceAll(value, " ", "")
				} else {
					value = strings.TrimSpace(value)
				}

				if len(row) == 2 {
					switch label {
					case "CPU":
						result, err := strconv.ParseFloat(numberReg.FindString(value), 64)
						if err != nil {
							lc.logger.Warnf("get lotus miner worker cpu err: %v", err)
						} else {
							ch <- prometheus.MustNewConstMetric(
								lc.workerCpuUse,
								prometheus.GaugeValue,
								result,
								minerNumber,
								daemonVersion,
								minerVersion,
								workerHostname,
								workerId,
							)
						}
					case "RAM":
						rams := ramvmemReg.FindAllString(value, -1)
						if len(rams) == 3 {
							useRate := rams[0]
							useBytes := DescSizeStr(rams[1])
							totalBytes := DescSizeStr(rams[2])

							useRateFloat, err := strconv.ParseFloat(useRate, 64)
							if err != nil {
								lc.logger.Warnf("get lotus miner worker ram err: %v", err)
							} else {
								ch <- prometheus.MustNewConstMetric(
									lc.workerRam,
									prometheus.GaugeValue,
									useRateFloat,
									minerNumber,
									daemonVersion,
									minerVersion,
									workerHostname,
									workerId,
									"rate",
								)
							}

							ch <- prometheus.MustNewConstMetric(
								lc.workerRam,
								prometheus.GaugeValue,
								float64(useBytes),
								minerNumber,
								daemonVersion,
								minerVersion,
								workerHostname,
								workerId,
								"used",
							)

							ch <- prometheus.MustNewConstMetric(
								lc.workerRam,
								prometheus.GaugeValue,
								float64(totalBytes),
								minerNumber,
								daemonVersion,
								minerVersion,
								workerHostname,
								workerId,
								"total",
							)
						}
					case "VMEM":
						rams := ramvmemReg.FindAllString(value, -1)
						if len(rams) == 3 {
							useRate := rams[0]
							useBytes := DescSizeStr(rams[1])
							totalBytes := DescSizeStr(rams[2])

							useRateFloat, err := strconv.ParseFloat(useRate, 64)
							if err != nil {
								lc.logger.Warnf("get lotus miner worker ram err: %v", err)
							} else {
								ch <- prometheus.MustNewConstMetric(
									lc.workerVmem,
									prometheus.GaugeValue,
									useRateFloat,
									minerNumber,
									daemonVersion,
									minerVersion,
									workerHostname,
									workerId,
									"rate",
								)
							}

							ch <- prometheus.MustNewConstMetric(
								lc.workerVmem,
								prometheus.GaugeValue,
								float64(useBytes),
								minerNumber,
								daemonVersion,
								minerVersion,
								workerHostname,
								workerId,
								"used",
							)

							ch <- prometheus.MustNewConstMetric(
								lc.workerVmem,
								prometheus.GaugeValue,
								float64(totalBytes),
								minerNumber,
								daemonVersion,
								minerVersion,
								workerHostname,
								workerId,
								"total",
							)
						}
					case "GPU":
						gpudata := strings.Split(value, ",")
						if len(gpudata) == 2 {
							gpuModel := gpudata[0]
							if strings.EqualFold(gpudata[1], " not used") {
								ch <- prometheus.MustNewConstMetric(
									lc.workerGpuIsUse,
									prometheus.GaugeValue,
									0,
									minerNumber,
									daemonVersion,
									minerVersion,
									workerHostname,
									workerId,
									gpuModel,
								)
							} else {
								ch <- prometheus.MustNewConstMetric(
									lc.workerGpuIsUse,
									prometheus.GaugeValue,
									1,
									minerNumber,
									daemonVersion,
									minerVersion,
									workerHostname,
									workerId,
									gpuModel,
								)
							}
						}
					}
				}
			}
		}
	}
	return nil
}

var numberReg = regexp.MustCompile(`\d+`)
var ramvmemReg = regexp.MustCompile(`\d+\.*\d*[TGiB]*`)
