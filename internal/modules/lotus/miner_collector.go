package lotus

import (
	"bytes"
	"fildr-cli/internal/gateway"
	"fildr-cli/internal/log"
	"github.com/fatih/color"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/build"
	"github.com/filecoin-project/lotus/chain/types"
	miner2 "github.com/filecoin-project/specs-actors/actors/builtin/miner"
	"github.com/filecoin-project/specs-actors/actors/builtin/power"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type minerCollector struct {
	miner  *miner
	logger log.Logger
}

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
		float64(types.BigMul(types.NewInt(secCounts.Active), types.NewInt(uint64(m.miner.sectorSizeUint))).Uint64()),
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

	return nil
}
