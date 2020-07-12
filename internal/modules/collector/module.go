package collector

import (
	"context"
	"fildr-cli/internal/config"
	"fildr-cli/internal/module"
	"fildr-cli/internal/modules/collector/metric/node"
	"fmt"
	"time"
)

var _ module.Module = (*Collector)(nil)

type Collector struct {
}

func New(ctx context.Context, config *config.TomlConfig) (*Collector, error) {
	//sysType := runtime.GOOS
	//if sysType == "linux" {
	for ns := range config.Collectors {
		if ns == "node" {
			metrics := config.Collectors["node"].Metric
			for i := range metrics {
				if metrics[i] == "cpu" {
					RegisterCollector(ns, "cpu", node.NewCpuCollector)
				}
			}
			fmt.Println(ns)
		}
	}
	//}
	return &Collector{}, nil
}

func (c *Collector) Name() string {
	return "collector"
}

func (c *Collector) Start() error {
	go func() {
		instance := GetInstance("node")
		instance.SetJob("test")
		instance.SetInstance("aaabc")
		for range time.Tick(time.Second * 10) {
			fmt.Println("print metrics ...")
			fmt.Println(instance.GetMetrics())
		}
	}()
	return nil
}

func (c *Collector) Stop() {

}
