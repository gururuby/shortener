package config

type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
}

func NewConfig() *Config {
	return &Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080",
	}
}
