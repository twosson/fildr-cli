package lotus

import (
	"context"
	"fmt"
)

type miner struct {
	minerNumber  string
	minerVersion string
	apiVersion   string

	lotusApi    *LotusApi
	lotusClient *LotusClient
	isShutdown  bool
}

func newMiner() (*miner, error) {
	miner := &miner{}
	miner.lotusApi = &LotusApi{}
	miner.lotusClient = &LotusClient{}
	miner.isShutdown = false
	if err := miner.lotusClient.WithMinerClient(miner.lotusApi); err != nil {
		return nil, err
	}

	minerVersion, apiVersion, err := miner.version()
	if err != nil {
		return nil, err
	}

	miner.minerVersion = minerVersion
	miner.apiVersion = apiVersion

	cancelCtx, cancel := context.WithCancel(context.Background())
	closingCh, err := miner.lotusApi.Closing(cancelCtx)
	if err != nil {
		cancel()
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
	addr, err := m.lotusApi.ActorAddress()
	if err != nil {
		return "", err
	}
	if addr.Empty() {
		return "", fmt.Errorf("get miner number empty")
	}
	return addr.String(), nil
}

func (m *miner) version() (string, string, error) {
	v, err := m.lotusApi.Version()
	if err != nil {
		return "", "", err
	}
	return v.Version, v.APIVersion.String(), nil
}
