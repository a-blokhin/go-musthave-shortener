package config

import (
	"flag"
)

type Config struct {
	ServerAddress string
	BaseURL       string
}

func ParseFlags() *Config {
	serverAddress := flag.String("a", "localhost:8080", "server address")
	baseURL := flag.String("b", "http://localhost:8080", "base URL for shortened links")

	flag.Parse()

	return &Config{
		ServerAddress: *serverAddress,
		BaseURL:       *baseURL,
	}
}
