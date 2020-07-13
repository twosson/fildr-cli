package node

import (
	"context"
	"fildr-cli/internal/config"
	"fildr-cli/internal/log"
	"fildr-cli/internal/module"
	"os"
	"time"
)

var _ module.Module = (*NodeCollectorModule)(nil)

type NodeCollectorModule struct {
	config *config.TomlConfig
}

func New(ctx context.Context, config *config.TomlConfig) (*NodeCollectorModule, error) {
	return &NodeCollectorModule{config: config}, nil
}

func (c *NodeCollectorModule) Name() string {
	return "collector"
}

func (c *NodeCollectorModule) Start() error {
	evaluation := c.config.Gateway.Evaluation
	if evaluation == 0 {
		evaluation = 5
	}

	instance := c.config.Gateway.Instance
	if instance == "" {
		hostname, err := os.Hostname()
		if err != nil {
			return err
		}
		instance = hostname
	}

	log.NopLogger().Infof("Start Node Collector ", instance, evaluation, c.config.Gateway.Url, c.config.Gateway.Token)
	c.execute(c.config.Gateway.Url, c.config.Gateway.Token, "node", instance, time.Duration(evaluation))

	return nil
}

func (c *NodeCollectorModule) Stop() {

}

func (c *NodeCollectorModule) execute(gateway string, token string, job string, instanceName string, evaluation time.Duration) {
	go func() {
		instance, err := GetInstance()
		if err != nil {
			log.NopLogger().Errorf("create node instance err", err)
			return
		}
		instance.SetJob(job)
		instance.SetInstance(instanceName)
		for range time.Tick(time.Second * evaluation) {
			metries, err := instance.GetMetrics()
			if err != nil {
				log.NopLogger().Named("collector-node").Errorf("instance get metrics err: ", err)
			}
			log.NopLogger().Infof("metries", metries)
			instance.PushMetrics(gateway, token, metries)
		}
	}()
}
