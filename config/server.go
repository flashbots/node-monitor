package config

type Server struct {
	ListenAddress string `yaml:"listen_address"`
	Name          string `yaml:"name"`
}
