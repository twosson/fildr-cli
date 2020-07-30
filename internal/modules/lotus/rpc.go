package lotus

import (
	"context"
	"fildr-cli/internal/config"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/go-jsonrpc/auth"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/sector-storage/fsutil"
	"github.com/filecoin-project/sector-storage/stores"
	"github.com/filecoin-project/sector-storage/storiface"
	"github.com/filecoin-project/specs-actors/actors/abi"
	miner2 "github.com/filecoin-project/specs-actors/actors/builtin/miner"
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"net/http"
)

type LotusApi struct {
	// MethodGroup: Auth
	AuthVerify func(token string) ([]auth.Permission, error)
	AuthNew    func(perms []auth.Permission) ([]byte, error)

	// MethodGroup: Net
	NetConnectedness func(peer.ID) (network.Connectedness, error)
	NetPeers         func() ([]peer.AddrInfo, error)
	NetConnect       func(peer.AddrInfo) error
	NetAddrsListen   func() (peer.AddrInfo, error)
	NetDisconnect    func(peer.ID) error
	NetFindPeer      func(peer.ID) (peer.AddrInfo, error)
	NetPubsubScores  func() ([]api.PubsubScore, error)

	// MethodGroup: Common

	// ID returns peerID of libp2p node backing this API
	ID func() (peer.ID, error)
	// Version provides information about API provider
	Version func() (api.Version, error)

	LogList     func() ([]string, error)
	LogSetLevel func(string, string) error

	// trigger graceful shutdown
	Shutdown func() error
	Closing  func(context.Context) (<-chan struct{}, error)

	SyncState func() (*api.SyncState, error)

	// Miner
	ActorAddress          func() (address.Address, error)
	StateMinerInfo        func(address.Address, types.TipSetKey) (api.MinerInfo, error)
	StateMinerPower       func(address.Address, types.TipSetKey) (*api.MinerPower, error)
	StateMinerSectorCount func(address.Address, types.TipSetKey) (api.MinerSectors, error)
	StateGetActor         func(actor address.Address, tsk types.TipSetKey) (*types.Actor, error)

	ChainReadObj func(cid.Cid) ([]byte, error)

	WalletBalance func(address.Address) (types.BigInt, error)

	StateMarketBalance func(address.Address, types.TipSetKey) (api.MarketBalance, error)

	SectorsList   func() ([]abi.SectorNumber, error)
	SectorsStatus func(sid abi.SectorNumber, showOnChainInfo bool) (api.SectorInfo, error)

	MarketListIncompleteDeals func() ([]storagemarket.MinerDeal, error)

	ChainHead            func() (*types.TipSet, error)
	StateMinerDeadlines  func(address.Address, types.TipSetKey) ([]*miner2.Deadline, error)
	StateMinerPartitions func(address.Address, uint64, types.TipSetKey) ([]*miner2.Partition, error)

	StorageList  func() (map[stores.ID][]stores.Decl, error)
	StorageLocal func() (map[stores.ID]string, error)
	StorageStat  func(id stores.ID) (fsutil.FsStat, error)
	StorageInfo  func(stores.ID) (stores.StorageInfo, error)

	WorkerJobs  func() (map[uint64][]storiface.WorkerJob, error)
	WorkerStats func() (map[uint64]storiface.WorkerStats, error)
}

type LotusClient struct {
	api      *LotusApi
	shutdown func()
}

var lotusMergeClient *LotusMergeClient

type LotusMergeClient struct {
	daemonClient *LotusClient
	minerClient  *LotusClient
}

func getLotusMergeClient() (*LotusMergeClient, error) {
	if lotusMergeClient == nil {
		lotusMergeClient = &LotusMergeClient{}
		daemonClient := &LotusClient{}
		if err := daemonClient.WithDaemonClient(); err != nil {
			return nil, err
		}
		minerClient := &LotusClient{}
		if err := minerClient.WithMinerClient(); err != nil {
			return nil, err
		}
		lotusMergeClient.daemonClient = daemonClient
		lotusMergeClient.minerClient = minerClient
	}
	return lotusMergeClient, nil
}

func (c *LotusClient) WithCommonClient(ctx context.Context, lotusApi *LotusApi, listenAddress string, token string) error {
	c.Shutdown()
	requestHeader := http.Header{}
	requestHeader.Add("Content-Type", "application/json")
	if token != "" {
		requestHeader.Add("Authorization", "Bearer "+token)
	}
	closer, err := jsonrpc.NewClient("ws://"+listenAddress+"/rpc/v0", "Filecoin", lotusApi, requestHeader)
	if err != nil {
		return err
	}
	c.shutdown = closer
	return nil
}

func (c *LotusClient) WithMinerClient() error {
	c.Shutdown()
	cfg := config.Get()
	requestHeader := http.Header{}
	requestHeader.Add("Content-Type", "application/json")
	token := cfg.Lotus.Miner.Token
	if len(token) > 10 {
		requestHeader.Add("Authorization", "Bearer "+token)
	}
	lotusApi := &LotusApi{}
	closer, err := jsonrpc.NewClient("ws://"+cfg.Lotus.Miner.ListenAddress+"/rpc/v0", "Filecoin", lotusApi, requestHeader)
	if err != nil {
		return err
	}
	c.shutdown = closer
	c.api = lotusApi
	return nil
}

func (c *LotusClient) WithDaemonClient() error {
	c.Shutdown()
	cfg := config.Get()
	requestHeader := http.Header{}
	requestHeader.Add("Content-Type", "application/json")
	token := cfg.Lotus.Daemon.Token
	if len(token) > 10 {
		requestHeader.Add("Authorization", "Bearer "+token)
	}

	lotusApi := &LotusApi{}
	closer, err := jsonrpc.NewClient("ws://"+cfg.Lotus.Daemon.ListenAddress+"/rpc/v0", "Filecoin", lotusApi, requestHeader)
	if err != nil {
		return err
	}
	c.shutdown = closer
	c.api = lotusApi
	return nil
}

func (c *LotusClient) Shutdown() {
	if c.shutdown != nil {
		c.shutdown()
	}
}
