package collector

import "fildr-cli/internal/modules/collector/metric/node"

func init() {
	registerCollector("node", "cpu", node.NewCpuCollector)
}
