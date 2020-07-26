package config

type Lotus struct {
	Daemon Daemon `mapstructure:"daemon"`
	Miner  Miner  `mapstructure:"miner"`
}

type Daemon struct {
	Enable        bool   `mapstructure:"enable"`
	ListenAddress string `mapstructure:"listen-address"`
	Token         string `mapstructure:"token"`
}

type Miner struct {
	Enable        bool   `mapstructure:"enable"`
	Path          string `mapstructure:"path"`
	ListenAddress string `mapstructure:"listen-address"`
	Token         string `mapstructure:"token"`
}
