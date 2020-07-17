package lotus

import (
	"context"
	"fildr-cli/internal/log"
	"fildr-cli/internal/module"
)

var _ module.Module = (*LotusCollectorModule)(nil)

type LotusCollectorModule struct {
	logger log.Logger
}

func New(ctx context.Context) (*LotusCollectorModule, error) {
	logger := log.From(ctx)
	return &LotusCollectorModule{logger: logger}, nil
}

func (l *LotusCollectorModule) Name() string {
	return `lotus-collector`
}

func (l *LotusCollectorModule) Start() error {
	panic("implement me")
}

func (l *LotusCollectorModule) Stop() {
	panic("implement me")
}
