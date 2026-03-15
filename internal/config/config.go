// Package config provides configuration management for the URL shortener service.
// It supports configuration from JSON config file, command-line flags and environment variables.
// Priority (lowest to highest): config file < flags < environment variables
package config

import (
	"log"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Config holds the configuration for the URL shortener service.
// It can be populated from command-line flags and environment variables.
type Config struct {
	ServerAddress     string `mapstructure:"server_address"`
	GRPCServerAddress string `mapstructure:"grpc_server_address"`
	BaseURL           string `mapstructure:"base_url"`
	FileStoragePath   string `mapstructure:"file_storage_path"`
	DatabaseDSN       string `mapstructure:"database_dsn"`
	DeleteURLs        DeleteURLsConfig
	AuditFile         string `mapstructure:"audit_file"`
	AuditURL          string `mapstructure:"audit_url"`
	EnableHTTPS       bool   `mapstructure:"enable_https"`
	SSLCertPath       string `mapstructure:"ssl_cert_path"`
	SSLKeyPath        string `mapstructure:"ssl_key_path"`
	TrustedSubnet     string `mapstructure:"trusted_subnet"`
}

// DeleteURLsConfig holds configuration for the asynchronous URL deletion feature.
type DeleteURLsConfig struct {
	BufferSize    int           `mapstructure:"buffer_size"`
	BatchSize     int           `mapstructure:"batch_size"`
	FlushInterval time.Duration `mapstructure:"flush_interval"`
	WorkerCount   int           `mapstructure:"worker_count"`
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
//
//	{
//	  "server_address": "localhost:8080",
//	  "base_url": "http://localhost",
//	  "file_storage_path": "/path/to/file.db",
//	  "database_dsn": "",
//	  "enable_https": true
//	}
//
// Returns a populated Config instance.
func ParseConfig() *Config {
	v := viper.New()

	setDefaults(v)

	bindFlags(v)

	readConfigFile(v)

	readEnvironmentVariables(v)

	config := &Config{}
	if err := v.Unmarshal(config); err != nil {
		log.Fatalf("Failed to unmarshal config: %v", err)
	}

	return config
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	v.SetDefault("server_address", "localhost:8080")
	v.SetDefault("grpc_server_address", "localhost:9090")
	v.SetDefault("base_url", "http://localhost:8080")
	v.SetDefault("file_storage_path", "")
	v.SetDefault("database_dsn", "")
	v.SetDefault("enable_https", false)
	v.SetDefault("ssl_cert_path", "cert.pem")
	v.SetDefault("ssl_key_path", "key.pem")
	v.SetDefault("audit_file", "")
	v.SetDefault("audit_url", "")
	v.SetDefault("trusted_subnet", "")
}

// bindFlags binds command-line flags to viper
func bindFlags(v *viper.Viper) {
	pflag.String("c", "", "path to JSON config file")
	pflag.String("config", "", "path to JSON config file")
	pflag.String("a", v.GetString("server_address"), "server address")
	pflag.String("g", v.GetString("grpc_server_address"), "gRPC server address")
	pflag.String("b", v.GetString("base_url"), "base URL for shortened links")
	pflag.String("f", v.GetString("file_storage_path"), "path to file storage (JSON format)")
	pflag.String("d", v.GetString("database_dsn"), "database connection string")
	pflag.Bool("s", v.GetBool("enable_https"), "enable HTTPS")
	pflag.String("cert-path", v.GetString("ssl_cert_path"), "path to SSL certificate file")
	pflag.String("key-path", v.GetString("ssl_key_path"), "path to SSL private key file")
	pflag.String("audit-file", v.GetString("audit_file"), "path to audit log file")
	pflag.String("audit-url", v.GetString("audit_url"), "URL of remote audit server")
	pflag.String("t", v.GetString("trusted_subnet"), "trusted subnet CIDR")

	pflag.Parse()

	if err := v.BindPFlags(pflag.CommandLine); err != nil {
		log.Printf("Failed to bind flags: %v", err)
	}
}

// readConfigFile reads configuration from JSON file
func readConfigFile(v *viper.Viper) {
	configPath := v.GetString("c")
	if configPath == "" {
		configPath = v.GetString("config")
	}
	if configPath == "" {
		configPath = v.GetString("CONFIG")
	}

	if configPath != "" {
		v.SetConfigFile(configPath)
		v.SetConfigType("json")

		if err := v.ReadInConfig(); err != nil {
			log.Printf("Failed to read config file %s: %v", configPath, err)
		}
	}
}

// readEnvironmentVariables reads configuration from environment variables
func readEnvironmentVariables(v *viper.Viper) {
	v.AutomaticEnv()

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	v.BindEnv("server_address", "SERVER_ADDRESS")
	v.BindEnv("grpc_server_address", "GRPC_SERVER_ADDRESS")
	v.BindEnv("base_url", "BASE_URL")
	v.BindEnv("file_storage_path", "FILE_STORAGE_PATH")
	v.BindEnv("database_dsn", "DATABASE_DSN")
	v.BindEnv("enable_https", "ENABLE_HTTPS")
	v.BindEnv("ssl_cert_path", "SSL_CERT_PATH")
	v.BindEnv("ssl_key_path", "SSL_KEY_PATH")
	v.BindEnv("audit_file", "AUDIT_FILE")
	v.BindEnv("audit_url", "AUDIT_URL")
	v.BindEnv("trusted_subnet", "TRUSTED_SUBNET")

	v.BindEnv("delete_urls.buffer_size", "DELETE_URLS_BUFFER_SIZE")
	v.BindEnv("delete_urls.batch_size", "DELETE_URLS_BATCH_SIZE")
	v.BindEnv("delete_urls.flush_interval", "DELETE_URLS_FLUSH_INTERVAL")
	v.BindEnv("delete_urls.worker_count", "DELETE_URLS_WORKER_COUNT")
}
