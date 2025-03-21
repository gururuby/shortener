package config

type Config struct {
	ServerAddress string
	PublicAddress string
}

func NewConfig() *Config {
	return &Config{
		ServerAddress: "localhost:8080",
		PublicAddress: "localhost:8080",
	}
}
