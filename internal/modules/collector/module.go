package collector

import (
	"context"
	"fil-pusher/internal/module"
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
	return nil
}

func (c *Collector) Stop() {

}
