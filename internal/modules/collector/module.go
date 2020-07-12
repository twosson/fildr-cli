package collector

import (
	"context"
	"fildr-cli/internal/module"
	"fmt"
	"time"
)

type Options struct {
}

var _ module.Module = (*Collector)(nil)

type Collector struct {
}

func New(ctx context.Context, options Options) (*Collector, error) {
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
