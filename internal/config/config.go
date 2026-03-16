// Package config provides configuration management for the URL shortener service.
// It supports configuration from JSON config file, command-line flags and environment variables.
// Priority (lowest to highest): config file < flags < environment variables
package config

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"time"

	"github.com/caarlos0/env/v11"
)

// Config holds the configuration for the URL shortener service.
// It can be populated from command-line flags and environment variables.
type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
	DeleteURLs      DeleteURLsConfig
	AuditFile       string `env:"AUDIT_FILE"`
	AuditURL        string `env:"AUDIT_URL"`
	EnableHTTPS     bool   `env:"ENABLE_HTTPS"`
	SSLCertPath     string `env:"SSL_CERT_PATH"`
	SSLKeyPath      string `env:"SSL_KEY_PATH"`
}

// DeleteURLsConfig holds configuration for the asynchronous URL deletion feature.
type DeleteURLsConfig struct {
	BufferSize    int           `env:"DELETE_URLS_BUFFER_SIZE"`
	BatchSize     int           `env:"DELETE_URLS_BATCH_SIZE"`
	FlushInterval time.Duration `env:"DELETE_URLS_FLUSH_INTERVAL"`
	WorkerCount   int           `env:"DELETE_URLS_WORKER_COUNT"`
}

// ParseConfig creates a new Config instance by parsing JSON config file,
// command-line flags and environment variables.
// Priority (lowest to highest): config file < flags < environment variables.
//
// Supported flags:
//   - c/-config: path to JSON config file
//   - a: server address (default: "localhost:8080")
//   - b: base URL for shortened links (default: "http://localhost:8080")
//   - f: path to file storage (JSON format)
//   - d: database connection string
//   - s: enable HTTPS (default: false)
//   - cert-path: path to SSL certificate file
//   - key-path: path to SSL private key file
//   - audit-file: path to audit log file
//   - audit-url: URL of remote audit server
//
// Supported environment variables:
//   - CONFIG: path to JSON config file
//   - SERVER_ADDRESS: server address
//   - BASE_URL: base URL for shortened links
//   - FILE_STORAGE_PATH: path to file storage
//   - DATABASE_DSN: database connection string
//   - ENABLE_HTTPS: enable HTTPS (true/false)
//   - SSL_CERT_PATH: path to SSL certificate file
//   - SSL_KEY_PATH: path to SSL private key file
//   - AUDIT_FILE: path to audit log file
//   - AUDIT_URL: URL of remote audit server
//   - DELETE_URLS_BUFFER_SIZE: buffer size for async deletion
//   - DELETE_URLS_BATCH_SIZE: batch size for async deletion
//   - DELETE_URLS_FLUSH_INTERVAL: flush interval for async deletion
//   - DELETE_URLS_WORKER_COUNT: worker count for async deletion
//
// JSON config file format:
//   {
//     "server_address": "localhost:8080",
//     "base_url": "http://localhost",
//     "file_storage_path": "/path/to/file.db",
//     "database_dsn": "",
//     "enable_https": true
//   }
//
// Returns a populated Config instance.
func ParseConfig() *Config {
	config := getDefaultConfig()

	if !flag.Parsed() {
		config = applyFlags(config)
	}

	applyEnvironmentVariables(config)

	return config
}

// getDefaultConfig returns a Config instance with default values
func getDefaultConfig() *Config {
	return &Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "",
		DatabaseDSN:     "",
		EnableHTTPS:     false,
		SSLCertPath:     "cert.pem",
		SSLKeyPath:      "key.pem",
		AuditFile:       "",
		AuditURL:        "",
	}
}

// applyFlags parses command-line flags and applies them to the config
func applyFlags(config *Config) *Config {
	configFlag := flag.String("c", "", "path to JSON config file")
	flag.StringVar(configFlag, "config", "", "path to JSON config file")
	serverAddressFlag := flag.String("a", config.ServerAddress, "server address")
	baseURLFlag := flag.String("b", config.BaseURL, "base URL for shortened links")
	fileStoragePathFlag := flag.String("f", config.FileStoragePath, "path to file storage (JSON format)")
	databaseDSNFlag := flag.String("d", config.DatabaseDSN, "database connection string")
	enableHTTPSFlag := flag.Bool("s", config.EnableHTTPS, "enable HTTPS")
	sslCertPathFlag := flag.String("cert-path", config.SSLCertPath, "path to SSL certificate file")
	sslKeyPathFlag := flag.String("key-path", config.SSLKeyPath, "path to SSL private key file")
	auditFileFlag := flag.String("audit-file", config.AuditFile, "path to audit log file")
	auditURLFlag := flag.String("audit-url", config.AuditURL, "URL of remote audit server")
	flag.Parse()

	// Load config from JSON file (lowest priority)
	configPath := *configFlag
	if configPath == "" {
		configPath = os.Getenv("CONFIG")
	}
	if configPath != "" {
		if fileConfig := loadConfigFromPath(configPath); fileConfig != nil {
			config = fileConfig
		}
	}

	config.ServerAddress = *serverAddressFlag
	config.BaseURL = *baseURLFlag
	config.FileStoragePath = *fileStoragePathFlag
	config.DatabaseDSN = *databaseDSNFlag
	config.EnableHTTPS = *enableHTTPSFlag
	config.SSLCertPath = *sslCertPathFlag
	config.SSLKeyPath = *sslKeyPathFlag
	config.AuditFile = *auditFileFlag
	config.AuditURL = *auditURLFlag

	return config
}

// applyEnvironmentVariables applies environment variables to the config
func applyEnvironmentVariables(config *Config) {
	envConfig := &Config{}
	if err := env.Parse(envConfig); err != nil {
		log.Printf("Failed to parse environment variables: %v", err)
		return
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
	if envConfig.AuditFile != "" {
		config.AuditFile = envConfig.AuditFile
	}
	if envConfig.AuditURL != "" {
		config.AuditURL = envConfig.AuditURL
	}
	if envConfig.EnableHTTPS {
		config.EnableHTTPS = envConfig.EnableHTTPS
	}
	if envConfig.SSLCertPath != "" {
		config.SSLCertPath = envConfig.SSLCertPath
	}
	if envConfig.SSLKeyPath != "" {
		config.SSLKeyPath = envConfig.SSLKeyPath
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
}

func loadConfigFromPath(path string) *Config {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("Failed to read config file %s: %v", path, err)
		return nil
	}

	var fileConfig struct {
		ServerAddress   string `json:"server_address"`
		BaseURL         string `json:"base_url"`
		FileStoragePath string `json:"file_storage_path"`
		DatabaseDSN     string `json:"database_dsn"`
		EnableHTTPS     bool   `json:"enable_https"`
	}

	if err := json.Unmarshal(data, &fileConfig); err != nil {
		log.Printf("Failed to parse config file %s: %v", path, err)
		return nil
	}

	return &Config{
		ServerAddress:   fileConfig.ServerAddress,
		BaseURL:         fileConfig.BaseURL,
		FileStoragePath: fileConfig.FileStoragePath,
		DatabaseDSN:     fileConfig.DatabaseDSN,
		EnableHTTPS:     fileConfig.EnableHTTPS,
		SSLCertPath:     "cert.pem",
		SSLKeyPath:      "key.pem",
		AuditFile:       "",
		AuditURL:        "",
	}
}
