package lotus

import (
	"context"
	"fmt"
	"github.com/filecoin-project/go-jsonrpc/auth"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMinerConn(t *testing.T) {
	lotusApi := &LotusApi{}
	lotusClient := &LotusClient{}
	err := lotusClient.WithCommonClient(context.Background(), lotusApi, "33s06815j0.qicp.vip:2345", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJyZWFkIiwid3JpdGUiLCJzaWduIiwiYWRtaW4iXX0.nprGvXwmDWPu3SzeYTsObLu75gmZ4a-LE-LHl4ysHjM")
	assert.NoError(t, err)
	lotusClient.Shutdown()
}

func TestDaemonConn(t *testing.T) {
	lotusApi := &LotusApi{}
	lotusClient := &LotusClient{}
	err := lotusClient.WithCommonClient(context.Background(), lotusApi, "127.0.0.1:1234", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJyZWFkIiwid3JpdGUiLCJzaWduIiwiYWRtaW4iXX0.VkP-gk6FOIcFIFOY4cifk_sn-N5_OqUJo2s7mWnzcfE")
	assert.NoError(t, err)
	lotusClient.Shutdown()
}

func TestAuthVerify(t *testing.T) {
	lotusApi := &LotusApi{}
	lotusClient := &LotusClient{}
	err := lotusClient.WithCommonClient(context.Background(), lotusApi, "127.0.0.1:1234", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJyZWFkIiwid3JpdGUiLCJzaWduIiwiYWRtaW4iXX0.VkP-gk6FOIcFIFOY4cifk_sn-N5_OqUJo2s7mWnzcfE")
	assert.NoError(t, err)
	permissions, err := lotusApi.AuthVerify("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJyZWFkIiwid3JpdGUiLCJzaWduIiwiYWRtaW4iXX0.VkP-gk6FOIcFIFOY4cifk_sn-N5_OqUJo2s7mWnzcfE")
	assert.NoError(t, err)
	assert.EqualValues(t, []auth.Permission{"read", "write", "sign", "admin"}, permissions)
	lotusClient.Shutdown()
}

func TestNetPeers(t *testing.T) {
	lotusApi := &LotusApi{}
	lotusClient := &LotusClient{}
	err := lotusClient.WithCommonClient(context.Background(), lotusApi, "127.0.0.1:1234", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJyZWFkIiwid3JpdGUiLCJzaWduIiwiYWRtaW4iXX0.VkP-gk6FOIcFIFOY4cifk_sn-N5_OqUJo2s7mWnzcfE")
	assert.NoError(t, err)
	peers, err := lotusApi.NetPeers()
	assert.NoError(t, err)
	for _, peer := range peers {
		fmt.Println(peer)
	}
	lotusClient.Shutdown()

}

func TestClient(t *testing.T) {

	lotusApi := &LotusApi{}
	lotusClient := &LotusClient{}
	err := lotusClient.WithCommonClient(context.Background(), lotusApi, "127.0.0.1:1234", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJyZWFkIiwid3JpdGUiLCJzaWduIiwiYWRtaW4iXX0.VkP-gk6FOIcFIFOY4cifk_sn-N5_OqUJo2s7mWnzcfE")
	assert.NoError(t, err)

	state, err := lotusApi.SyncState()
	assert.NoError(t, err)
	for _, sync := range state.ActiveSyncs {
		fmt.Println(sync.Base.Key().String())
		fmt.Println(sync.Target.Key().String())
		fmt.Println(sync.Stage)
		fmt.Println(sync.Height)
		fmt.Println(sync.End)
		fmt.Println(sync.Target.Height() - sync.Height)
	}

	//client := &Client{}
	//requestHeader := http.Header{}
	//requestHeader.Add("Content-Type", "application/json")
	////requestHeader.Add("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJyZWFkIiwid3JpdGUiLCJzaWduIiwiYWRtaW4iXX0.Ss-Hzt-aqlS1XkWbOKqfR0I79mIn8KkBlOCI29uW3CE")
	//closer, err := jsonrpc.NewClient("ws://193.112.74.252:1234/rpc/v0", "Filecoin", client, requestHeader)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//defer closer()
	//ch, err := client.MpoolSub()
	//if err != nil {
	//	fmt.Println(err)
	//}
	//for {
	//	select {
	//	case a := <- ch:
	//		fmt.Println(a)
	//	}
	//}

	//var cl struct {
	//	MpoolSub func(context.Context) (<-chan api.MpoolUpdate, error)
	//}
	//
	//requestHeader := http.Header{}
	//requestHeader.Add("Content-Type", "application/json")
	//clo, err := jsonrpc.NewClient("ws://193.112.74.252:1234/rpc/v0", "Filecoin", &cl, requestHeader)
	//defer clo()
	//if err != nil {
	//	fmt.Println(err)
	//}
	//ctx, cancel := context.WithCancel(context.Background())
	//defer cancel()
	//
	//sub, err := cl.MpoolSub(ctx)
	//for  {
	//	select {
	//	case a := <- sub:
	//		fmt.Println(a.Message.Message.From)
	//	case <-time.After(time.Second * 3):
	//		fmt.Println("3s...")
	//	}
	//}

	/**
	curl -X POST \
	     -H "Content-Type: application/json" \
	     --data '{ "jsonrpc": "2.0", "method": "Filecoin.StateMinerPower", "params": ["t0121737", null], "id": 3 }' \
	     'http://193.112.74.252:1234/rpc/v0'
	*/

}
