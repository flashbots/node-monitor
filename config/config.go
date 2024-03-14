package config

type Config struct {
	Eth    Eth    `yaml:"eth"`
	Log    Log    `yaml:"log"`
	Server Server `yaml:"server"`
}
