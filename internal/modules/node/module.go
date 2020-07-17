package node

import (
	"context"
	"fildr-cli/internal/config"
	"fildr-cli/internal/log"
	"fildr-cli/internal/module"
	"fildr-cli/internal/pusher"
	"time"
)

var _ module.Module = (*NodeCollectorModule)(nil)

var (
	namespace = "node"
	factories = make(map[string]func(logger log.Logger) (pusher.Collector, error))
)

func registerCollector(collector string, factory func(logger log.Logger) (pusher.Collector, error)) {
	factories[collector] = factory
}

type NodeCollectorModule struct {
	promInstance *pusher.PromInstance
	logger       log.Logger
}

func New(ctx context.Context) (*NodeCollectorModule, error) {
	logger := log.From(ctx)

	fc, err := pusher.NewFildrCollector(ctx, namespace)
	if err != nil {
		return nil, err
	}

	for k, c := range factories {
		collector, err := c(logger)
		if err != nil {
			logger.Warnf("collector %s is err: %v", k, err)
			continue
		}
		fc.Registry(k, collector)
	}

	promInstance, err := pusher.GetPromInstance(ctx, namespace, fc)
	if err != nil {
		return nil, err
	}

	return &NodeCollectorModule{logger: logger, promInstance: promInstance}, nil
}

func (mod *NodeCollectorModule) Name() string {
	return "node-collector"
}

func (mod *NodeCollectorModule) Start() error {
	mod.logger.Infof("node collector starting ...")

	cfg := config.Get()
	eval := cfg.Gateway.Evaluation

	go func() {
		for range time.Tick(eval) {
			metric, err := mod.promInstance.GetMetrics()
			if err != nil {
				mod.logger.Warnf("get metrics err: %v", err)
				continue
			}

			if err = mod.promInstance.PushMetrics(metric); err != nil {
				mod.logger.Warnf("push metrics err: %v", err)
			}
		}
	}()

	return nil
}

func (mod *NodeCollectorModule) Stop() {

}
