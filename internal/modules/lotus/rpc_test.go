package lotus

import (
	"testing"
)

func TestClient(t *testing.T) {
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
