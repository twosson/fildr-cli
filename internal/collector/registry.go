package collector

import "fil-pusher/internal/collector/metric/node"

func init() {
	registerCollector("node", "cpu", node.NewCpuCollector)
}
