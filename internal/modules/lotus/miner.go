package lotus

import (
	"context"
	"fmt"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/chain/types"
)

type miner struct {
	minerNumber    string
	sectorSize     string
	sectorSizeUint uint64
	ownerNumber    string
	peerId         string
	minerVersion   string
	apiVersion     string

	lotusClient *LotusMergeClient
	isShutdown  bool
}

func newMiner() (*miner, error) {
	miner := &miner{}

	lotusClient, err := getLotusMergeClient()
	if err != nil {
		return nil, err
	}

	miner.lotusClient = lotusClient
	miner.isShutdown = false

	minerVersion, apiVersion, err := miner.version()
	if err != nil {
		return nil, err
	}

	miner.minerVersion = minerVersion
	miner.apiVersion = apiVersion

	cancelCtx, cancel := context.WithCancel(context.Background())
	closingCh, err := miner.lotusClient.minerClient.api.Closing(cancelCtx)
	if err != nil {
		cancel()
		return nil, err
	}

	minerNumber, err := miner.getMinerNumber()
	if err != nil {
		return nil, err
	}

	miner.minerNumber = minerNumber

	sectorSize, sectorSizeUint, ownerNumber, peerId, err := miner.getStateMinerInfo()
	miner.sectorSize = sectorSize
	miner.ownerNumber = ownerNumber
	miner.peerId = peerId
	miner.sectorSizeUint = sectorSizeUint
	if err != nil {
		return nil, err
	}

	go func() {
		<-closingCh
		miner.isShutdown = true
		cancel()
	}()

	return miner, nil
}

func (m *miner) getMinerNumber() (string, error) {
	addr, err := m.lotusClient.minerClient.api.ActorAddress()
	if err != nil {
		return "", err
	}
	if addr.Empty() {
		return "", fmt.Errorf("get miner number empty")
	}
	return addr.String(), nil
}

func (m *miner) getStateMinerInfo() (sectorSize string, sectorSizeUint uint64, ownerNumber string, peerId string, err error) {
	addr, err := address.NewFromString(m.minerNumber)
	if err != nil {
		return
	}
	info, err := m.lotusClient.daemonClient.api.StateMinerInfo(addr, types.EmptyTSK)
	if err != nil {
		return
	}
	ownerNumber = info.Owner.String()
	peerId = info.PeerId.String()
	sectorSize = info.SectorSize.ShortString()
	sectorSizeUint = uint64(info.SectorSize)
	return
}

func (m *miner) version() (string, string, error) {
	v, err := m.lotusClient.minerClient.api.Version()
	if err != nil {
		return "", "", err
	}
	return v.Version, v.APIVersion.String(), nil
}
