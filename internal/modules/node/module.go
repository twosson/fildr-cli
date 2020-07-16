package node

import (
	"context"
	"fildr-cli/internal/config"
	"fildr-cli/internal/log"
	"fildr-cli/internal/module"
	"time"
)

var _ module.Module = (*NodeCollectorModule)(nil)

type NodeCollectorModule struct {
	logger log.Logger
}

func New(ctx context.Context) (*NodeCollectorModule, error) {
	logger := log.From(ctx)
	return &NodeCollectorModule{logger: logger}, nil
}

func (c *NodeCollectorModule) Name() string {
	return "collector"
}

func (c *NodeCollectorModule) Start() error {
	c.logger.Infof("node collector starting ...")

	gateway := config.Get().Gateway
	instance, err := GetInstance(c.logger)
	if err != nil {
		return err
	}
	instance.SetInstance(gateway.Instance)
	instance.SetJob("node")

	go func() {
		for range time.Tick(gateway.Evaluation) {
			metric, err := instance.GetMetrics()
			if err != nil {
				c.logger.Warnf("get metrics err: %v", err)
				continue
			}

			if err = instance.PushMetrics(gateway.Url, gateway.Token, metric); err != nil {
				c.logger.Warnf("push metrics err: %v", err)
			}
		}
	}()

	return nil
}

func (c *NodeCollectorModule) Stop() {

}
