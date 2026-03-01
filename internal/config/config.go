package config

import (
	"flag"
	"log"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
	DeleteURLs      DeleteURLsConfig
}

type DeleteURLsConfig struct {
	BufferSize     int           `env:"DELETE_URLS_BUFFER_SIZE"`
	BatchSize      int           `env:"DELETE_URLS_BATCH_SIZE"`
	FlushInterval  time.Duration `env:"DELETE_URLS_FLUSH_INTERVAL"`
	WorkerCount    int           `env:"DELETE_URLS_WORKER_COUNT"`
}

func ParseConfig() (config *Config) {

	defaultServerAddress := "localhost:8080"
	defaultBaseURL := "http://localhost:8080"
	defaultFileStoragePath := ""
	defaultDatabaseDSN := ""

	if !flag.Parsed() {
		serverAddressFlag := flag.String("a", defaultServerAddress, "server address")
		baseURLFlag := flag.String("b", defaultBaseURL, "base URL for shortened links")
		fileStoragePathFlag := flag.String("f", defaultFileStoragePath, "path to file storage (JSON format)")
		databaseDSNFlag := flag.String("d", defaultDatabaseDSN, "database connection string")
		flag.Parse()

		config = &Config{
			ServerAddress:   *serverAddressFlag,
			BaseURL:         *baseURLFlag,
			FileStoragePath: *fileStoragePathFlag,
			DatabaseDSN:     *databaseDSNFlag,
		}
	} else {
		config = &Config{
			ServerAddress:   defaultServerAddress,
			BaseURL:         defaultBaseURL,
			FileStoragePath: defaultFileStoragePath,
			DatabaseDSN:     defaultDatabaseDSN,
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
	if envConfig.DatabaseDSN != "" {
		config.DatabaseDSN = envConfig.DatabaseDSN
	}

	if envConfig.DeleteURLs.BufferSize != 0 {
		config.DeleteURLs.BufferSize = envConfig.DeleteURLs.BufferSize
	}
	if envConfig.DeleteURLs.BatchSize != 0 {
		config.DeleteURLs.BatchSize = envConfig.DeleteURLs.BatchSize
	}
	if envConfig.DeleteURLs.FlushInterval != 0 {
		config.DeleteURLs.FlushInterval = envConfig.DeleteURLs.FlushInterval
	}
	if envConfig.DeleteURLs.WorkerCount != 0 {
		config.DeleteURLs.WorkerCount = envConfig.DeleteURLs.WorkerCount
	}

	return config
}
