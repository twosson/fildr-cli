package main

func main() {

	//var client struct {
	//	ChainHead        func() (*types.TipSet, error)
	//	WalletBalance    func(address address.Address) (types.BigInt, error)
	//	StateNetworkName func() (dtypes.NetworkName, error)
	//	ChainNotify      func() (<-chan []*api.HeadChange, error)
	//	SyncState        func() (*api.SyncState, error)
	//}
	//
	//requestHeader := http.Header{}
	//
	//requestHeader.Add("Content-Type", "application/json")
	//
	//closer, err := jsonrpc.NewClient("ws://193.112.74.252:1234/rpc/v0", "Filecoin", &client, requestHeader)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//
	//a, err := client.ChainHead()
	//if err != nil {
	//	fmt.Println(err)
	//}
	//fmt.Println(a.Height().String())
	//
	//addr, err := address.NewFromString("t1c72pztmp4jmw2v6lbhyu25uxozb6udn4z6dsegi")
	//if err != nil {
	//	fmt.Println(err)
	//}
	//
	//bl, err := client.WalletBalance(addr)
	//fmt.Println(bl)
	//
	//c, err := client.StateNetworkName()
	//fmt.Println(c)
	//
	//ch, err := client.ChainNotify()
	//if err != nil {
	//	fmt.Println(err)
	//}
	//for {
	//	fmt.Println(ch)
	//}
	//
	//defer closer()

}
