package lotus

import (
	"fmt"
	"github.com/filecoin-project/lotus/build"
	"github.com/filecoin-project/lotus/chain/types"
	ma "github.com/multiformats/go-multiaddr"
	"math/big"
	"strconv"
	"strings"
)

func addrsToIpAndPort(addrs []ma.Multiaddr) (ip string, port string, err error) {
	if addrs == nil || len(addrs) == 0 {
		err = fmt.Errorf("net peer addrs is nil")
		return
	}

	addr := addrs[0].String()
	splits := strings.Split(addr, "/")
	if len(splits) == 5 {
		ip = splits[2]
		port = splits[4]
		return
	} else {
		err = fmt.Errorf("net peer addrs parsing error")
	}
	return
}

func bigIntToFil(num types.BigInt) (float64, error) {
	fil := new(big.Rat).SetFrac(num.Int, big.NewInt(int64(build.FilecoinPrecision)))
	if fil.Sign() == 0 {
		return 0, nil
	}
	filStr := strings.TrimRight(strings.TrimRight(fil.FloatString(18), "0"), ".")
	return strconv.ParseFloat(filStr, 64)
}
