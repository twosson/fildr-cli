package lotus

import (
	"fildr-cli/internal/config"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/sector-storage/stores"
	"github.com/filecoin-project/sector-storage/storiface"
	"github.com/filecoin-project/specs-actors/actors/abi"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"net/http"
	"strconv"
)

type Client struct {
	ID      func() (peer.ID, error)
	Version func() (api.Version, error)
	LogList func() ([]string, error)

	// Full
	MpoolSub func() (<-chan api.MpoolUpdate, error)

	// Common
	NetConnectedness func(peer.ID) (network.Connectedness, error)
	NetPeers         func() ([]peer.AddrInfo, error)
	NetConnect       func(peer.AddrInfo) error
	NetAddrsListen   func() (peer.AddrInfo, error)
	NetDisconnect    func(peer.ID) error
	NetFindPeer      func(peer.ID) (peer.AddrInfo, error)
	NetPubsubScores  func() ([]api.PubsubScore, error)
	LogSetLevel      func(string, string) error

	// Miner
	ActorAddress    func() (address.Address, error)
	ActorSectorSize func(address.Address) (abi.SectorSize, error)
	MiningBase      func() (*types.TipSet, error)
	PledgeSector    func() error
	SectorsStatus   func() (api.SectorInfo, error)
	SectorsList     func() ([]abi.SectorNumber, error)
	SectorsRefs     func() (map[string][]api.SealedRef, error)
	SectorsUpdate   func(abi.SectorNumber, api.SectorState) error
	SectorRemove    func(abi.SectorNumber) error
	StorageList     func() (map[stores.ID][]stores.Decl, error)
	StorageLocal    func() (map[stores.ID]string, error)
	StorageStat     func(stores.ID) (stores.FsStat, error)
	WorkerConnect   func(string) error
	WorkerStats     func() (map[uint64]storiface.WorkerStats, error)
}

func InitClient(client *Client) (jsonrpc.ClientCloser, error) {
	cfg := config.Get()
	requestHeader := http.Header{}
	requestHeader.Add("Content-Type", "application/json")
	ip := cfg.Lotus.Daemon.Ip
	port := cfg.Lotus.Daemon.Port
	return jsonrpc.NewClient("ws://"+ip+":"+strconv.Itoa(port)+"/rpc/v0", "Filecoin", client, requestHeader)
}
