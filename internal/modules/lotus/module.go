package lotus

import (
	"context"
	"fildr-cli/internal/config"
	"fildr-cli/internal/log"
	"fildr-cli/internal/module"
)

var _ module.Module = (*LotusCollectorModule)(nil)

type LotusCollectorModule struct {
	roles  []string
	config *config.TomlConfig
	logger log.Logger
}

func New(ctx context.Context, config *config.TomlConfig) (*LotusCollectorModule, error) {
	logger := log.From(ctx)
	return &LotusCollectorModule{config: config, logger: logger}, nil
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
