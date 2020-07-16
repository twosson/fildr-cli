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

	viper.Set("gateway.url", "https://api.fildr.com/fildr-miner")
	viper.Set("gateway.token", "")
	viper.Set("gateway.instance", "")
	viper.Set("gateway.evaluation", "5s")

	viper.Set("lotus.daemon.enable", false)
	viper.Set("lotus.daemon.ip", "127.0.0.1")
	viper.Set("lotus.daemon.port", 1234)
	viper.Set("lotus.daemon.token", "")

	return viper.WriteConfig()
}
