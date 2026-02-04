package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
}

func ParseConfig() (config *Config) {

	defaultServerAddress := "localhost:8080"
	defaultBaseURL := "http://localhost:8080"
	defaultFileStoragePath := ""

	if !flag.Parsed() {
		serverAddressFlag := flag.String("a", defaultServerAddress, "server address")
		baseURLFlag := flag.String("b", defaultBaseURL, "base URL for shortened links")
		fileStoragePathFlag := flag.String("f", defaultFileStoragePath, "path to file storage (JSON format)")
		flag.Parse()

		config = &Config{
			ServerAddress:   *serverAddressFlag,
			BaseURL:         *baseURLFlag,
			FileStoragePath: *fileStoragePathFlag,
		}
	} else {
		config = &Config{
			ServerAddress:   defaultServerAddress,
			BaseURL:         defaultBaseURL,
			FileStoragePath: defaultFileStoragePath,
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
	if envConfig.FileStoragePath != "" {
		config.FileStoragePath = envConfig.FileStoragePath
	}

	return config
}
