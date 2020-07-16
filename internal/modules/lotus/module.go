package lotus

import (
	"context"
	"fildr-cli/internal/config"
	"fildr-cli/internal/log"
	"fildr-cli/internal/module"
)

var _ module.Module = (*LotusCollectorModule)(nil)

type LotusCollectorModule struct {
	config *config.TomlConfig
	logger log.Logger
}

func New(ctx context.Context, config *config.TomlConfig) (*LotusCollectorModule, error) {

	return nil, nil
}

func (l *LotusCollectorModule) Name() string {
	panic("implement me")
}

func (l *LotusCollectorModule) Start() error {
	panic("implement me")
}

func (l *LotusCollectorModule) Stop() {
	panic("implement me")
}
