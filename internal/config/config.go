package config

import (
	"github.com/spf13/viper"
	"os"
	"os/user"
)

type Config struct {
	Gateway Gateway `mapstructure:"gateway"`
	Lotus   Lotus   `mapstructure:"lotus"`
}

var cfg = Config{}

// 加载配置文件
func LoadConfig() error {
	user, err := user.Current()
	if err != nil {
		return err
	}
	path := user.HomeDir + "/.fildr/config.toml"
	viper.SetConfigType("toml")
	viper.SetConfigFile(path)

	if err = viper.ReadInConfig(); err != nil {
		return err
	}
	return viper.Unmarshal(&cfg)
}

func Get() Config {
	if cfg.Gateway.Instance == "" {
		hostname, err := os.Hostname()
		if err != nil {
			hostname = "unknown"
		}
		cfg.Gateway.Instance = hostname
	}

	return cfg
}

func InitializationConfig() error {
	user, err := user.Current()
	if err != nil {
		return err
	}
	path := user.HomeDir + "/.fildr"
	_, err = os.Stat(path)
	if err != nil {
		err = os.Mkdir(path, os.ModePerm)
		if err != nil {
			return err
		}
	}

	viper.SetConfigType("toml")
	viper.SetConfigFile(path + `/config.toml`)

	viper.Set("gateway.url", viper.GetString("gateway.url"))
	viper.Set("gateway.token", viper.GetString("gateway.token"))
	viper.Set("gateway.instance", viper.GetString("gateway.instance"))
	viper.Set("gateway.evaluation", viper.GetDuration("gateway.evaluation"))

	viper.Set("lotus.daemon.enable", false)
	viper.Set("lotus.daemon.listen-address", "127.0.0.1:1234")
	viper.Set("lotus.daemon.token", "")

	viper.Set("lotus.miner.enable", false)
	viper.Set("lotus.miner.path", "")

	return viper.WriteConfig()
}
