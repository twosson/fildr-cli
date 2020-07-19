package node

import (
	"context"
	"fildr-cli/internal/gateway"
	"fildr-cli/internal/log"
	"fildr-cli/internal/module"
)

var _ module.Module = (*NodeCollectorModule)(nil)

var (
	namespace = "node"
	factories = make(map[string]func(logger log.Logger) (gateway.Collector, error))
)

func registerCollector(collector string, factory func(logger log.Logger) (gateway.Collector, error)) {
	factories[collector] = factory
}

type NodeCollectorModule struct {
	logger log.Logger
}

func New(ctx context.Context) (*NodeCollectorModule, error) {
	logger := log.From(ctx)
	return &NodeCollectorModule{logger: logger}, nil
}

func (mod *NodeCollectorModule) Name() string {
	return "node-collector"
}

func (mod *NodeCollectorModule) Start() error {
	mod.logger.Infof("node collector starting ...")
	for k, c := range factories {
		collector, err := c(mod.logger)
		if err != nil {
			mod.logger.Warnf("collector %s is err: %v", k, err)
			continue
		}
		gateway.Registry("node", k, collector)
	}
	return nil
}

func (mod *NodeCollectorModule) Stop() {

}
