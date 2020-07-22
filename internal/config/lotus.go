package config

type Lotus struct {
	Daemon Daemon `mapstructure:"daemon"`
	Miner  Miner  `mapstructure:"miner"`
}

type Daemon struct {
	Enable bool   `mapstructure:"enable"`
	Ip     string `mapstructure:"ip"`
	Port   int    `mapstructure:"port"`
}

type Miner struct {
	Enable bool   `mapstructure:"enable"`
	Path   string `mapstructure:"path"`
}
