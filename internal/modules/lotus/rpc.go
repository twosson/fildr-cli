package lotus

import (
	"fildr-cli/internal/config"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/lotus/api"
	"github.com/libp2p/go-libp2p-core/peer"
	"net/http"
)

type Client struct {
	Version         func() (api.Version, error)
	NetPeers        func() ([]peer.AddrInfo, error)
	NetPubsubScores func() ([]api.PubsubScore, error)
}

func InitClient(client *Client) (jsonrpc.ClientCloser, error) {
	cfg := config.Get()
	requestHeader := http.Header{}
	requestHeader.Add("Content-Type", "application/json")
	ip := cfg.Lotus.Daemon.Ip
	port := cfg.Lotus.Daemon.Port
	return jsonrpc.NewClient("ws://"+ip+":"+string(port)+"/rpc/v0", "Filecoin", client, requestHeader)
}
