package lotus

import (
	"bytes"
	"fildr-cli/internal/gateway"
	"fildr-cli/internal/log"
	"github.com/fatih/color"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/build"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/sector-storage/storiface"
	miner2 "github.com/filecoin-project/specs-actors/actors/builtin/miner"
	"github.com/filecoin-project/specs-actors/actors/builtin/power"
	sealing "github.com/filecoin-project/storage-fsm"
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
	"strings"
	"time"
)

type minerCollector struct {
	miner  *miner
	logger log.Logger
}

var jobsCache map[uint64]storiface.WorkerJob = make(map[uint64]storiface.WorkerJob, 10)

func init() {
	registerCollector("lotus-miner", NewMinerCollector)
}

func NewMinerCollector(logger log.Logger) (gateway.Collector, error) {
	miner, err := newMiner()
	if err != nil {
		return nil, err
	}
	return &minerCollector{miner: miner, logger: logger}, nil
}

func (m *minerCollector) Update(ch chan<- prometheus.Metric) error {

	var runState float64 = 1
	if m.miner.isShutdown {
		runState = 0
	}

	ch <- prometheus.MustNewConstMetric(
		minerInfoDesc,
		prometheus.GaugeValue,
		runState,
		m.miner.ownerNumber,
		m.miner.minerNumber,
		m.miner.minerVersion,
		m.miner.apiVersion,
		m.miner.sectorSize,
		m.miner.peerId,
	)

	peers, err := m.miner.lotusClient.minerClient.api.NetPeers()
	if err != nil {
		m.logger.Warnf("Lotus miner net peers err: %v", err)
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		minerConnectionsDesc,
		prometheus.GaugeValue,
		float64(len(peers)),
		m.miner.ownerNumber,
		m.miner.minerNumber,
	)

	scores, err := m.miner.lotusClient.minerClient.api.NetPubsubScores()
	if err != nil {
		m.logger.Warnf("Lotus miner net pubsub scores err: %v", err)
		return err
	}

	scoresMap := make(map[string]float64, len(scores))
	for _, score := range scores {
		scoresMap[score.ID.Pretty()] = score.Score
	}

	for _, peer := range peers {

		ip, port, err := addrsToIpAndPort(peer.Addrs)
		if err != nil {
			continue
		}

		score, ok := scoresMap[peer.ID.Pretty()]
		if !ok {
			score = 0
		}

		ch <- prometheus.MustNewConstMetric(
			minerPeersDesc,
			prometheus.GaugeValue,
			score,
			m.miner.ownerNumber,
			m.miner.minerNumber,
			peer.ID.Pretty(),
			ip,
			port,
		)
	}

	addr, err := address.NewFromString(m.miner.minerNumber)
	if err != nil {
		return err
	}

	pow, err := m.miner.lotusClient.daemonClient.api.StateMinerPower(addr, types.EmptyTSK)
	if err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		minerBytePowerDesc,
		prometheus.GaugeValue,
		float64(pow.MinerPower.RawBytePower.Int64()),
		m.miner.ownerNumber,
		m.miner.minerNumber,
	)

	ch <- prometheus.MustNewConstMetric(
		totalBytePowerDesc,
		prometheus.GaugeValue,
		float64(pow.TotalPower.RawBytePower.Int64()),
		m.miner.ownerNumber,
		m.miner.minerNumber,
	)

	ch <- prometheus.MustNewConstMetric(
		minerActualPowerDesc,
		prometheus.GaugeValue,
		float64(pow.MinerPower.QualityAdjPower.Int64()),
		m.miner.ownerNumber,
		m.miner.minerNumber,
	)

	ch <- prometheus.MustNewConstMetric(
		totalActualPowerDesc,
		prometheus.GaugeValue,
		float64(pow.TotalPower.QualityAdjPower.Int64()),
		m.miner.ownerNumber,
		m.miner.minerNumber,
	)

	rpercI := types.BigDiv(types.BigMul(pow.MinerPower.RawBytePower, types.NewInt(1000000)), pow.TotalPower.RawBytePower)
	qpercI := types.BigDiv(types.BigMul(pow.MinerPower.QualityAdjPower, types.NewInt(1000000)), pow.TotalPower.QualityAdjPower)

	ch <- prometheus.MustNewConstMetric(
		bytePowerRateDesc,
		prometheus.GaugeValue,
		float64(rpercI.Int64())/10000,
		m.miner.ownerNumber,
		m.miner.minerNumber,
	)

	ch <- prometheus.MustNewConstMetric(
		actualPowerRateDesc,
		prometheus.GaugeValue,
		float64(qpercI.Int64())/10000,
		m.miner.ownerNumber,
		m.miner.minerNumber,
	)

	secCounts, err := m.miner.lotusClient.daemonClient.api.StateMinerSectorCount(addr, types.EmptyTSK)
	if err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		committedDesc,
		prometheus.GaugeValue,
		float64(types.BigMul(types.NewInt(secCounts.Sectors), types.NewInt(uint64(m.miner.sectorSizeUint))).Uint64()),
		m.miner.ownerNumber,
		m.miner.minerNumber,
	)

	ch <- prometheus.MustNewConstMetric(
		provingDesc,
		prometheus.GaugeValue,
		float64(types.BigMul(types.NewInt(secCounts.Sectors), types.NewInt(uint64(m.miner.sectorSizeUint))).Uint64()),
		m.miner.ownerNumber,
		m.miner.minerNumber,
	)

	if pow.MinerPower.RawBytePower.LessThan(power.ConsensusMinerMinPower) {
		ch <- prometheus.MustNewConstMetric(
			blockWinRateDesc,
			prometheus.GaugeValue,
			0,
			m.miner.ownerNumber,
			m.miner.minerNumber,
		)
		ch <- prometheus.MustNewConstMetric(
			blockWinPerDayDesc,
			prometheus.GaugeValue,
			0,
			m.miner.ownerNumber,
			m.miner.minerNumber,
		)
	} else {
		expWinChance := float64(types.BigMul(qpercI, types.NewInt(build.BlocksPerEpoch)).Int64()) / 1000000
		if expWinChance > 0 {
			if expWinChance > 1 {
				expWinChance = 1
			}
			winRate := time.Duration(float64(time.Second*time.Duration(build.BlockDelaySecs)) / expWinChance)
			winPerDay := float64(time.Hour*24) / float64(winRate)
			color.Blue("%.4f/day (every %s)", winPerDay, winRate.Truncate(time.Second))
			ch <- prometheus.MustNewConstMetric(
				blockWinRateDesc,
				prometheus.GaugeValue,
				winRate.Seconds(),
				m.miner.ownerNumber,
				m.miner.minerNumber,
			)
			ch <- prometheus.MustNewConstMetric(
				blockWinPerDayDesc,
				prometheus.GaugeValue,
				winPerDay,
				m.miner.ownerNumber,
				m.miner.minerNumber,
			)
		} else {
			ch <- prometheus.MustNewConstMetric(
				blockWinRateDesc,
				prometheus.GaugeValue,
				0,
				m.miner.ownerNumber,
				m.miner.minerNumber,
			)
			ch <- prometheus.MustNewConstMetric(
				blockWinPerDayDesc,
				prometheus.GaugeValue,
				0,
				m.miner.ownerNumber,
				m.miner.minerNumber,
			)
		}
	}

	mact, err := m.miner.lotusClient.daemonClient.api.StateGetActor(addr, types.EmptyTSK)
	if err != nil {
		return err
	}

	minerBalance, err := bigIntToFil(mact.Balance)
	if err != nil {
		return err
	}
	ch <- prometheus.MustNewConstMetric(
		minerBalanceDesc,
		prometheus.GaugeValue,
		minerBalance,
		m.miner.ownerNumber,
		m.miner.minerNumber,
	)

	var mas miner2.State
	{
		rmas, err := m.miner.lotusClient.daemonClient.api.ChainReadObj(mact.Head)
		if err != nil {
			return err
		}
		if err := mas.UnmarshalCBOR(bytes.NewReader(rmas)); err != nil {
			return err
		}
	}

	minerBalancePreCommit, err := bigIntToFil(mas.PreCommitDeposits)
	if err != nil {
		return err
	}
	ch <- prometheus.MustNewConstMetric(
		minerBalancePreCommitDesc,
		prometheus.GaugeValue,
		minerBalancePreCommit,
		m.miner.ownerNumber,
		m.miner.minerNumber,
	)

	minerBalanceLocked, err := bigIntToFil(mas.LockedFunds)
	if err != nil {
		return err
	}
	ch <- prometheus.MustNewConstMetric(
		minerBalanceLockedDesc,
		prometheus.GaugeValue,
		minerBalanceLocked,
		m.miner.ownerNumber,
		m.miner.minerNumber,
	)

	minerBalanceAvailable, err := bigIntToFil(types.BigSub(mact.Balance, types.BigAdd(mas.LockedFunds, mas.PreCommitDeposits)))
	if err != nil {
		return err
	}
	ch <- prometheus.MustNewConstMetric(
		minerBalanceAvailableDesc,
		prometheus.GaugeValue,
		minerBalanceAvailable,
		m.miner.ownerNumber,
		m.miner.minerNumber,
	)

	mi, err := m.miner.lotusClient.daemonClient.api.StateMinerInfo(addr, types.EmptyTSK)
	if err != nil {
		return err
	}

	wb, err := m.miner.lotusClient.daemonClient.api.WalletBalance(mi.Worker)
	if err != nil {
		return err
	}

	wbfil, err := bigIntToFil(wb)
	if err != nil {
		return err
	}
	ch <- prometheus.MustNewConstMetric(
		workerBalanceDesc,
		prometheus.GaugeValue,
		wbfil,
		m.miner.ownerNumber,
		m.miner.minerNumber,
	)

	mb, err := m.miner.lotusClient.daemonClient.api.StateMarketBalance(addr, types.EmptyTSK)
	if err != nil {
		return err
	}

	escrow, err := bigIntToFil(mb.Escrow)
	if err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		marketEscrowBalanceDesc,
		prometheus.GaugeValue,
		escrow,
		m.miner.ownerNumber,
		m.miner.minerNumber,
	)

	locked, err := bigIntToFil(mb.Locked)
	if err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		marketLockedBalanceDesc,
		prometheus.GaugeValue,
		locked,
		m.miner.ownerNumber,
		m.miner.minerNumber,
	)

	sectors, err := m.miner.lotusClient.minerClient.api.SectorsList()
	if err != nil {
		return err
	}

	buckets := map[sealing.SectorState]int{
		"Total":                len(sectors),
		"Proving":              0,
		"WaitDeals":            0,
		"UndefinedSectorState": 0,
		"Empty":                0,
		"Packing":              0,
		"PreCommit1":           0,
		"PreCommit2":           0,
		"PreCommitting":        0,
		"PreCommitWait":        0,
		"WaitSeed":             0,
		"Committing":           0,
		"CommitWait":           0,
		"FinalizeSector":       0,
		"FailedUnrecoverable":  0,
		"SealPreCommit1Failed": 0,
		"SealPreCommit2Failed": 0,
		"PreCommitFailed":      0,
		"ComputeProofFailed":   0,
		"CommitFailed":         0,
		"PackingFailed":        0,
		"FinalizeFailed":       0,
		"Faulty":               0,
		"FaultReported":        0,
		"FaultedFinal":         0,
	}

	for _, s := range sectors {
		st, err := m.miner.lotusClient.minerClient.api.SectorsStatus(s, false)
		if err != nil {
			m.logger.Warnf("lotus get sectors status err: %v", err)
			return err
		}
		buckets[sealing.SectorState(st.State)]++
	}

	for state, i := range buckets {
		na := strings.ToLower(string(state))
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, minerNamespace, "ss"+na),
				"Filecoin sector state for "+na,
				[]string{ownerNumber, minerNumber},
				nil,
			),
			prometheus.GaugeValue,
			float64(i),
			m.miner.ownerNumber,
			m.miner.minerNumber,
		)
	}

	deals, err := m.miner.lotusClient.minerClient.api.MarketListIncompleteDeals()
	if err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		dealsDesc,
		prometheus.GaugeValue,
		float64(len(deals)),
		m.miner.ownerNumber,
		m.miner.minerNumber,
	)

	head, err := m.miner.lotusClient.daemonClient.api.ChainHead()
	if err != nil {
		return err
	}

	deadlines, err := m.miner.lotusClient.daemonClient.api.StateMinerDeadlines(addr, head.Key())
	if err != nil {
		return err
	}

	parts := map[uint64][]*miner2.Partition{}
	for dlIdx := range deadlines {
		part, err := m.miner.lotusClient.daemonClient.api.StateMinerPartitions(addr, uint64(dlIdx), types.EmptyTSK)
		if err != nil {
			return err
		}

		parts[uint64(dlIdx)] = part
	}

	proving := uint64(0)
	faults := uint64(0)
	recovering := uint64(0)

	for _, partitions := range parts {
		for _, partition := range partitions {
			sc, err := partition.Sectors.Count()
			if err != nil {
				return err
			}
			proving += sc

			fc, err := partition.Faults.Count()
			if err != nil {
				return err
			}
			faults += fc

			rc, err := partition.Faults.Count()
			if err != nil {
				return err
			}
			recovering += rc
		}
	}

	ch <- prometheus.MustNewConstMetric(
		faultsDesc,
		prometheus.GaugeValue,
		float64(faults),
		m.miner.ownerNumber,
		m.miner.minerNumber,
	)

	ch <- prometheus.MustNewConstMetric(
		recoveringDesc,
		prometheus.GaugeValue,
		float64(recovering),
		m.miner.ownerNumber,
		m.miner.minerNumber,
	)

	var faultPerc float64
	if proving > 0 {
		faultPerc = float64(faults*10000/proving) / 100
	}

	ch <- prometheus.MustNewConstMetric(
		faultPercDesc,
		prometheus.GaugeValue,
		faultPerc,
		m.miner.ownerNumber,
		m.miner.minerNumber,
	)

	st, err := m.miner.lotusClient.minerClient.api.StorageList()
	if err != nil {
		return err
	}

	local, err := m.miner.lotusClient.minerClient.api.StorageLocal()
	if err != nil {
		return err
	}

	for id, decls := range st {
		pingStart := time.Now()
		st, err := m.miner.lotusClient.minerClient.api.StorageStat(id)
		if err != nil {
			m.logger.Warnf("get storage state err: %v", err)
			continue
		}
		ping := time.Now().Sub(pingStart)

		var cnt [3]int
		for _, decl := range decls {
			for i := range cnt {
				if decl.SectorFileType&(1<<i) != 0 {
					cnt[i]++
				}
			}
		}

		si, err := m.miner.lotusClient.minerClient.api.StorageInfo(id)
		if err != nil {
			return err
		}

		localPath, _ := local[id]

		var url string
		if len(si.URLs) > 0 {
			url = si.URLs[0]
		}

		canSeal := "no"
		canStore := "no"
		if si.CanSeal {
			canSeal = "yes"
		}
		if si.CanStore {
			canStore = "yes"
		}

		ch <- prometheus.MustNewConstMetric(
			storageInfoDesc,
			prometheus.GaugeValue,
			1,
			m.miner.ownerNumber,
			m.miner.minerNumber,
			string(id),
			strconv.FormatUint(si.Weight, 10),
			canSeal,
			canStore,
			localPath,
			url,
		)

		usedPercent := (st.Capacity - st.Available) * 100 / st.Capacity

		ch <- prometheus.MustNewConstMetric(
			storageMetricsDesc,
			prometheus.GaugeValue,
			float64(usedPercent),
			m.miner.ownerNumber,
			m.miner.minerNumber,
			string(id),
			"usedpercent",
		)

		ch <- prometheus.MustNewConstMetric(
			storageMetricsDesc,
			prometheus.GaugeValue,
			float64(st.Capacity),
			m.miner.ownerNumber,
			m.miner.minerNumber,
			string(id),
			"capacity",
		)

		ch <- prometheus.MustNewConstMetric(
			storageMetricsDesc,
			prometheus.GaugeValue,
			float64(st.Available),
			m.miner.ownerNumber,
			m.miner.minerNumber,
			string(id),
			"available",
		)

		ch <- prometheus.MustNewConstMetric(
			storageMetricsDesc,
			prometheus.GaugeValue,
			float64(cnt[0]),
			m.miner.ownerNumber,
			m.miner.minerNumber,
			string(id),
			"unsealed",
		)

		ch <- prometheus.MustNewConstMetric(
			storageMetricsDesc,
			prometheus.GaugeValue,
			float64(cnt[1]),
			m.miner.ownerNumber,
			m.miner.minerNumber,
			string(id),
			"sealed",
		)

		ch <- prometheus.MustNewConstMetric(
			storageMetricsDesc,
			prometheus.GaugeValue,
			float64(cnt[2]),
			m.miner.ownerNumber,
			m.miner.minerNumber,
			string(id),
			"caches",
		)

		ch <- prometheus.MustNewConstMetric(
			storageMetricsDesc,
			prometheus.GaugeValue,
			float64(ping.Truncate(time.Microsecond*100).Microseconds()),
			m.miner.ownerNumber,
			m.miner.minerNumber,
			string(id),
			"latency",
		)

	}

	wst, err := m.miner.lotusClient.minerClient.api.WorkerStats()
	if err != nil {
		return err
	}

	jobs, err := m.miner.lotusClient.minerClient.api.WorkerJobs()
	if err != nil {
		m.logger.Warnf("lotus get worker jobs err: %v", err)
		return err
	}

	var tmpJobs map[uint64]storiface.WorkerJob = make(map[uint64]storiface.WorkerJob, 10)

	for wid, jobs := range jobs {
		for _, job := range jobs {
			tmpJobs[job.ID] = job
			jobsCache[job.ID] = job
			ch <- prometheus.MustNewConstMetric(
				jobsDesc,
				prometheus.GaugeValue,
				time.Now().Sub(job.Start).Truncate(time.Millisecond*100).Seconds(),
				m.miner.ownerNumber,
				m.miner.minerNumber,
				strconv.FormatUint(wid, 10),
				strconv.FormatUint(job.ID, 10),
				job.Sector.Number.String(),
				wst[wid].Info.Hostname,
				job.Task.Short(),
				job.Start.String(),
			)
		}
	}

	for jid, job := range jobsCache {
		if _, ok := tmpJobs[jid]; !ok {
			ch <- prometheus.MustNewConstMetric(
				jobsDurationDesc,
				prometheus.GaugeValue,
				time.Now().Sub(job.Start).Truncate(time.Millisecond*100).Seconds(),
				m.miner.ownerNumber,
				m.miner.minerNumber,
				job.Task.Short(),
			)
			delete(jobsCache, jid)
		}
	}

	for wid, stats := range wst {
		vmem := stats.Info.Resources.MemPhysical + stats.Info.Resources.MemSwap

		ch <- prometheus.MustNewConstMetric(
			workerInfoDesc,
			prometheus.GaugeValue,
			1,
			m.miner.ownerNumber,
			m.miner.minerNumber,
			strconv.FormatUint(wid, 10),
			stats.Info.Hostname,
			strconv.FormatUint(stats.Info.Resources.CPUs, 10),
			strconv.FormatUint(stats.Info.Resources.MemPhysical, 10),
			strconv.FormatUint(vmem, 10),
		)

		ch <- prometheus.MustNewConstMetric(
			workerMetricsDesc,
			prometheus.GaugeValue,
			float64(stats.CpuUse),
			m.miner.ownerNumber,
			m.miner.minerNumber,
			strconv.FormatUint(wid, 10),
			stats.Info.Hostname,
			"cpu_used",
		)

		ch <- prometheus.MustNewConstMetric(
			workerMetricsDesc,
			prometheus.GaugeValue,
			float64(stats.Info.Resources.MemReserved+stats.MemUsedMin),
			m.miner.ownerNumber,
			m.miner.minerNumber,
			strconv.FormatUint(wid, 10),
			stats.Info.Hostname,
			"ram_used",
		)

		ch <- prometheus.MustNewConstMetric(
			workerMetricsDesc,
			prometheus.GaugeValue,
			float64(stats.Info.Resources.MemReserved+stats.MemUsedMax),
			m.miner.ownerNumber,
			m.miner.minerNumber,
			strconv.FormatUint(wid, 10),
			stats.Info.Hostname,
			"vmem_used",
		)

		var gpuUsed float64 = 0
		if stats.GpuUsed {
			gpuUsed = 1
		}

		gpus := stats.Info.Resources.GPUs
		if gpus != nil && len(gpus) > 0 {
			for _, s := range gpus {
				ch <- prometheus.MustNewConstMetric(
					workerGpuDesc,
					prometheus.GaugeValue,
					gpuUsed,
					m.miner.ownerNumber,
					m.miner.minerNumber,
					strconv.FormatUint(wid, 10),
					stats.Info.Hostname,
					s,
				)
			}
		}
	}

	return nil
}
