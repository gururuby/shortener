package config

type Config struct {
	ServerAddress string
	ServerBaseURL string
}

func NewConfig() *Config {
	return &Config{
		ServerAddress: "localhost:8080",
		ServerBaseURL: "http://localhost:8080",
	}
}
