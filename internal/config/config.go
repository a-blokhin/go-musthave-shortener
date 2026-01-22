package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
}

func ParseConfig() (config *Config) {

	defaultServerAddress := "localhost:8080"
	defaultBaseURL := "http://localhost:8080"

	if !flag.Parsed() {
		serverAddressFlag := flag.String("a", defaultServerAddress, "server address")
		baseURLFlag := flag.String("b", defaultBaseURL, "base URL for shortened links")
		flag.Parse()

		config = &Config{
			ServerAddress: *serverAddressFlag,
			BaseURL:       *baseURLFlag,
		}
	} else {
		config = &Config{
			ServerAddress: defaultServerAddress,
			BaseURL:       defaultBaseURL,
		}
	}

	envConfig := &Config{}
	if err := env.Parse(envConfig); err != nil {
		log.Printf("Failed to parse environment variables: %v", err)
	}

	if envConfig.ServerAddress != "" {
		config.ServerAddress = envConfig.ServerAddress
	}
	if envConfig.BaseURL != "" {
		config.BaseURL = envConfig.BaseURL
	}

	return config
}
