package config

import "time"

type Gateway struct {
	Url        string        `mapstructure:"url"`
	Token      string        `mapstructure:"token"`
	Instance   string        `mapstructure:"instance"`
	Evaluation time.Duration `mapstructure:"evaluation"`
}
