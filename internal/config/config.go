// Package config provides configuration management for the URL shortener service.
// It supports configuration from command-line flags and environment variables.
package config

import (
	"flag"
	"log"
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

// ParseConfig creates a new Config instance by parsing command-line flags
// and environment variables. Environment variables take precedence over flags.
//
// Supported flags:
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
// Returns a populated Config instance.
func ParseConfig() (config *Config) {

	defaultServerAddress := "localhost:8080"
	defaultBaseURL := "http://localhost:8080"
	defaultFileStoragePath := ""
	defaultDatabaseDSN := ""
	defaultEnableHTTPS := false
	defaultSSLCertPath := "cert.pem"
	defaultSSLKeyPath := "key.pem"
	defaultAuditFile := ""
	defaultAuditURL := ""

	if !flag.Parsed() {
		serverAddressFlag := flag.String("a", defaultServerAddress, "server address")
		baseURLFlag := flag.String("b", defaultBaseURL, "base URL for shortened links")
		fileStoragePathFlag := flag.String("f", defaultFileStoragePath, "path to file storage (JSON format)")
		databaseDSNFlag := flag.String("d", defaultDatabaseDSN, "database connection string")
		enableHTTPSFlag := flag.Bool("s", defaultEnableHTTPS, "enable HTTPS")
		sslCertPathFlag := flag.String("cert-path", defaultSSLCertPath, "path to SSL certificate file")
		sslKeyPathFlag := flag.String("key-path", defaultSSLKeyPath, "path to SSL private key file")
		auditFileFlag := flag.String("audit-file", defaultAuditFile, "path to audit log file")
		auditURLFlag := flag.String("audit-url", defaultAuditURL, "URL of remote audit server")
		flag.Parse()

		config = &Config{
			ServerAddress:   *serverAddressFlag,
			BaseURL:         *baseURLFlag,
			FileStoragePath: *fileStoragePathFlag,
			DatabaseDSN:     *databaseDSNFlag,
			EnableHTTPS:     *enableHTTPSFlag,
			SSLCertPath:     *sslCertPathFlag,
			SSLKeyPath:      *sslKeyPathFlag,
			AuditFile:       *auditFileFlag,
			AuditURL:        *auditURLFlag,
		}
	} else {
		config = &Config{
			ServerAddress:   defaultServerAddress,
			BaseURL:         defaultBaseURL,
			FileStoragePath: defaultFileStoragePath,
			DatabaseDSN:     defaultDatabaseDSN,
			EnableHTTPS:     defaultEnableHTTPS,
			SSLCertPath:     defaultSSLCertPath,
			SSLKeyPath:      defaultSSLKeyPath,
			AuditFile:       defaultAuditFile,
			AuditURL:        defaultAuditURL,
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

	return config
}
