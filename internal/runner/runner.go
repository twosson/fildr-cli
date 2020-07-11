package runner

import (
	"context"
	"fil-pusher/internal/log"
)

type Options struct {
}

type Runner struct {
}

func NewRunner(ctx context.Context, logger log.Logger, options Options) (*Runner, error) {
	return &Runner{}, nil
}

func (r *Runner) Start(ctx context.Context, startupCh, shutdownCh chan bool) {
	logger := log.From(ctx)

	go func() {

	}()

	<-ctx.Done()

	shutdownCtx := log.WithLoggerContext(context.Background(), logger)
	r.Stop(shutdownCtx)
	shutdownCh <- true
}

func (r *Runner) Stop(ctx context.Context) {

}
