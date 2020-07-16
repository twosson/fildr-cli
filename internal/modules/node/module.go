package node

import (
	"context"
	"fildr-cli/internal/config"
	"fildr-cli/internal/log"
	"fildr-cli/internal/module"
	"fmt"
	"os"
	"time"
)

var _ module.Module = (*NodeCollectorModule)(nil)

type NodeCollectorModule struct {
	config *config.TomlConfig
	logger log.Logger
}

func New(ctx context.Context, config *config.TomlConfig) (*NodeCollectorModule, error) {
	logger := log.From(ctx)
	return &NodeCollectorModule{config: config, logger: logger}, nil
}

func (c *NodeCollectorModule) Name() string {
	return "collector"
}

func (c *NodeCollectorModule) Start() error {
	c.logger.Infof("a module for node collector started.")
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

	c.logger.Infof("node collector push gateway url : %s", c.config.Gateway.Url)
	c.logger.Infof("node collector evaluation time : %ds", evaluation)
	c.logger.Infof("node collector job : %s", "node")
	c.logger.Infof("node collector instance : %s", instance)
	c.execute(c.config.Gateway.Url, c.config.Gateway.Token, "node", instance, time.Duration(evaluation))

	return nil
}

func (c *NodeCollectorModule) Stop() {

}

func (c *NodeCollectorModule) execute(gateway string, token string, job string, instanceName string, evaluation time.Duration) {
	go func() {
		instance, err := GetInstance()
		if err != nil {
			fmt.Println("get instance err : ", err)
			return
		}
		instance.SetJob(job)
		instance.SetInstance(instanceName)

		for range time.Tick(time.Second * evaluation) {
			metries, err := instance.GetMetrics()
			if err != nil {
				c.logger.Errorf("instance get metrics err: %v", err.Error())
			}
			instance.PushMetrics(gateway, token, metries)
		}
	}()
}
