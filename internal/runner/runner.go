package runner

import (
	"context"
	"fildr-cli/internal/config"
	"fildr-cli/internal/gateway"
	"fildr-cli/internal/log"
	"fildr-cli/internal/module"
	"fildr-cli/internal/modules/lotus"
	"fildr-cli/internal/modules/node"
	"fmt"
)

type Options struct {
	Context string
}

type Runner struct {
	moduleManager *module.Manager
}

func NewRunner(ctx context.Context, logger log.Logger, options Options) (*Runner, error) {
	ctx = log.WithLoggerContext(ctx, logger)

	r := Runner{}

	if options.Context != "" {
		logger.With("initial-context", options.Context).Infof("Settiing initial context from user flags")
	}

	if err := config.LoadConfig(); err != nil {
		return nil, err
	}

	moduleManager, err := initModuleManager(logger)
	if err != nil {
		return nil, fmt.Errorf("init module manager: %w", err)
	}
	r.moduleManager = moduleManager

	moduleList, err := initModules(ctx)
	if err != nil {
		return nil, fmt.Errorf("initializing modules: %w", err)
	}

	for _, mod := range moduleList {
		if err := moduleManager.Register(mod); err != nil {
			return nil, fmt.Errorf("loading module %s: %w", mod.Name(), err)
		}
	}

	gateway.Run(ctx)

	return &r, nil
}

func (r *Runner) Start(ctx context.Context, startupCh, shutdownCh chan bool) {
	logger := log.From(ctx)

	if startupCh != nil {
		startupCh <- true
	}

	<-ctx.Done()

	shutdownCtx := log.WithLoggerContext(context.Background(), logger)
	r.Stop(shutdownCtx)
	shutdownCh <- true
}

func (r *Runner) Stop(ctx context.Context) {

}

func initModuleManager(logger log.Logger) (*module.Manager, error) {
	moduleManager, err := module.NewManager(logger)
	if err != nil {
		return nil, fmt.Errorf("create module manager: %w", err)
	}

	return moduleManager, nil
}

func initModules(ctx context.Context) ([]module.Module, error) {
	var list []module.Module

	nodeCollector, err := node.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("initialize node collector module: %v", err)
	}

	lotusCollector, err := lotus.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("initialize lotus collector module: %v", err)
	}

	list = append(list, nodeCollector)
	list = append(list, lotusCollector)

	return list, nil
}
