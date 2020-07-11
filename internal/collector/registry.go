package collector

import "fil-pusher/internal/metric/node"

func init() {
	registerCollector("node", "cpu", node.NewCpuCollector)
}
