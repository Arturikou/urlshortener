package config

import (
	"flag"
	"fmt"

	"github.com/Arturikou/urlshortener/internal/utils"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Server          ServerConfig
	LogLevel        string `env:"LOG_LEVEL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
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

func Load() (*Config, error) {
	cfg := &Config{}

	flag.StringVar(&cfg.Server.Addr, "a", "localhost:8080", "The address to listen on for HTTP requests.")
	flag.BoolVar(&cfg.Server.EnableHTTPS, "s", false, "Enable HTTPS")
	flag.StringVar(&cfg.LogLevel, "level", "info", "Log level")
	flag.StringVar(&cfg.FileStoragePath, "f", "storage.json", "File storage path")
	flag.StringVar(&cfg.Handlers.BaseAddr, "b", "http://localhost:8080", "The address for shortener response")
	flag.StringVar(&cfg.Database.DSN, "d", "", "Database DSN")
	flag.StringVar(&cfg.Audit.AuditFile, "audit-file", "", "AuditConfig file path")
	flag.StringVar(&cfg.Audit.AuditURL, "audit-url", "", "AuditConfig URL")
	flag.Parse()

	if err := cleanenv.ReadEnv(cfg); err != nil {
		return nil, fmt.Errorf("failed to read env variables: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
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
