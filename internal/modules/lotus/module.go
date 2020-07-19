package lotus

import (
	"context"
	"fildr-cli/internal/gateway"
	"fildr-cli/internal/log"
	"fildr-cli/internal/module"
)

var _ module.Module = (*LotusCollectorModule)(nil)

var (
	namespace = "lotus"
	factories = make(map[string]func(logger log.Logger) (gateway.Collector, error))
)

func registerCollector(collector string, factory func(logger log.Logger) (gateway.Collector, error)) {
	factories[collector] = factory
}

type LotusCollectorModule struct {
	logger log.Logger
}

func New(ctx context.Context) (*LotusCollectorModule, error) {
	logger := log.From(ctx)
	return &LotusCollectorModule{logger: logger}, nil
}

func (mod *LotusCollectorModule) Name() string {
	return `lotus-collector`
}

func (mod *LotusCollectorModule) Start() error {
	for k, c := range factories {
		collector, err := c(mod.logger)
		if err != nil {
			mod.logger.Warnf("collector %s is err: %v", k, err)
			continue
		}
		gateway.Registry("lotus", k, collector)
	}
	return nil
}

func (mod *LotusCollectorModule) Stop() {

}
