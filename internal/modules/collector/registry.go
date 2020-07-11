package collector

import "fil-pusher/internal/modules/collector/metric/node"

func init() {
	registerCollector("node", "cpu", node.NewCpuCollector)
}
