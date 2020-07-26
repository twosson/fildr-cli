package lotus

import (
	"fmt"
	ma "github.com/multiformats/go-multiaddr"
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
