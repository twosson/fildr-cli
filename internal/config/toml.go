package config

import (
	"github.com/BurntSushi/toml"
	"os"
	"path/filepath"
)

var (
	cfg *TomlConfig
)

func Config() (*TomlConfig, error) {
	if cfg != nil {
		return cfg, nil
	}
	filePath, err := filepath.Abs(os.Getenv("HOME") + "/.fildr/config.toml")
	if err != nil {
		return nil, err
	}
	if _, err := toml.DecodeFile(filePath, &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

type TomlConfig struct {
	Gateway    Gateway
	Collectors map[string]Collector
}

type Gateway struct {
	Url        string
	Token      string
	Instance   string
	Evaluation int64
}

type Collector struct {
	Metric []string
}
