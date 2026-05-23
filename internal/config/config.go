package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/Arturikou/urlshortener/internal/utils"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Server          ServerConfig
	LogLevel        string `env:"LOG_LEVEL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	ConfigPath      string `env:"CONFIG"`
	Handlers        HandlersConfig
	Database        DatabaseConfig
	Audit           AuditConfig
}

type ServerConfig struct {
	Addr        string `env:"SERVER_ADDRESS"`
	PprofAddr   string `env:"PPROF_ADDRESS" env-default:"localhost:6060"`
	EnableHTTPS bool   `env:"ENABLE_HTTPS"`
	CertFile    string `env:"CERT_FILE" env-default:"cert.pem"`
	KeyFile     string `env:"KEY_FILE" env-default:"key.pem"`
}

type HandlersConfig struct {
	BaseAddr string `env:"BASE_URL"`
}

type DatabaseConfig struct {
	DSN string `env:"DATABASE_DSN"`
}

type AuditConfig struct {
	AuditFile string `env:"AUDIT_FILE"`
	AuditURL  string `env:"AUDIT_URL"`
}

type jsonConfig struct {
	ServerAddress   string `json:"server_address"`
	BaseURL         string `json:"base_url"`
	FileStoragePath string `json:"file_storage_path"`
	DatabaseDSN     string `json:"database_dsn"`
	EnableHTTPS     bool   `json:"enable_https"`
}

// Load priority: ENV > flags > JSON > defaults
func Load() (*Config, error) {
	cfg := parseFlags()

	if err := cleanenv.ReadEnv(cfg); err != nil {
		return nil, fmt.Errorf("failed to read env variables: %w", err)
	}

	if err := applyJSONConfig(cfg); err != nil {
		return nil, fmt.Errorf("failed to apply json config: %w", err)
	}

	// Defaults:
	if cfg.Server.Addr == "" {
		cfg.Server.Addr = "localhost:8080"
	}

	if cfg.Handlers.BaseAddr == "" {
		cfg.Handlers.BaseAddr = "http://localhost:8080"
	}

	if cfg.FileStoragePath == "" {
		cfg.FileStoragePath = "storage.json"
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

func parseFlags() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.Server.Addr, "a", "", "The address to listen on for HTTP requests.")
	flag.BoolVar(&cfg.Server.EnableHTTPS, "s", false, "Enable HTTPS")
	flag.StringVar(&cfg.LogLevel, "level", "info", "Log level")
	flag.StringVar(&cfg.FileStoragePath, "f", "", "File storage path")
	flag.StringVar(&cfg.Handlers.BaseAddr, "b", "", "The address for shortener response")
	flag.StringVar(&cfg.Database.DSN, "d", "", "Database DSN")
	flag.StringVar(&cfg.Audit.AuditFile, "audit-file", "", "AuditConfig file path")
	flag.StringVar(&cfg.Audit.AuditURL, "audit-url", "", "AuditConfig URL")
	flag.StringVar(&cfg.ConfigPath, "c", "", "json config file path")
	flag.StringVar(&cfg.ConfigPath, "config", "", "json config file path")
	flag.Parse()

	return cfg
}

func applyJSONConfig(cfg *Config) error {
	if cfg.ConfigPath == "" {
		return nil
	}

	file, err := os.Open(cfg.ConfigPath)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer func() { _ = file.Close() }()

	j := &jsonConfig{}
	if err = json.NewDecoder(file).Decode(j); err != nil {
		return fmt.Errorf("decode: %w", err)
	}

	if cfg.Server.Addr == "" {
		cfg.Server.Addr = j.ServerAddress
	}

	if cfg.Handlers.BaseAddr == "" {
		cfg.Handlers.BaseAddr = j.BaseURL
	}

	if cfg.FileStoragePath == "" {
		cfg.FileStoragePath = j.FileStoragePath
	}

	if cfg.Database.DSN == "" {
		cfg.Database.DSN = j.DatabaseDSN
	}

	if j.EnableHTTPS {
		cfg.Server.EnableHTTPS = true
	}

	return nil
}

func (cfg *Config) validate() error {
	if cfg.FileStoragePath != "" {
		if err := utils.ValidateFilePath(cfg.FileStoragePath); err != nil {
			return fmt.Errorf("storage path: %v", err)
		}
	}

	if cfg.Audit.AuditFile != "" {
		if err := utils.ValidateFilePath(cfg.Audit.AuditFile); err != nil {
			return fmt.Errorf("audit file: %v", err)
		}
	}

	return nil
}
