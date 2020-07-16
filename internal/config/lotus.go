package config

type Lotus struct {
	Daemon Daemon `mapstructure:"daemon"`
}

type Daemon struct {
	Enable bool   `mapstructure:"enable"`
	Ip     string `mapstructure:"ip"`
	Port   int    `mapstructure:"port"`
	Token  string `mapstructure:"token"`
}
