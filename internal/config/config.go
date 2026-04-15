package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	ServerAddr      string `env:"SERVER_ADDRESS"`
	LogLevel        string `env:"LOG_LEVEL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Handlers        HandlersConfig
	Database        DatabaseConfig
	Audit           AuditConfig
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

func MustLoad() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.ServerAddr, "a", "localhost:8080", "The address to listen on for HTTP requests.")
	flag.StringVar(&cfg.LogLevel, "level", "info", "Log level")
	flag.StringVar(&cfg.FileStoragePath, "f", "storage.json", "File storage path")
	flag.StringVar(&cfg.Handlers.BaseAddr, "b", "http://localhost:8080", "The address for shortener response")
	flag.StringVar(&cfg.Database.DSN, "d", "", "Database DSN")
	flag.StringVar(&cfg.Audit.AuditFile, "audit-file", "", "AuditConfig file path")
	flag.StringVar(&cfg.Audit.AuditURL, "audit-url", "", "AuditConfig URL")
	flag.Parse()

	if err := cleanenv.ReadEnv(cfg); err != nil {
		log.Fatalf("failed to read env variables: %v", err)
	}

	if err := cfg.validate(); err != nil {
		log.Fatalf("config validation failed: %v", err)
	}

	return cfg
}

func (cfg *Config) validate() error {
	if cfg.FileStoragePath != "" {
		if err := validateFilePath(cfg.FileStoragePath); err != nil {
			return fmt.Errorf("storage path: %v", err)
		}
	}

	if cfg.Audit.AuditFile != "" {
		if err := validateFilePath(cfg.Audit.AuditFile); err != nil {
			return fmt.Errorf("audit file: %v", err)
		}
	}

	return nil
}

func validateFilePath(filePath string) error {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return err
	}
	dir := filepath.Dir(absPath)

	fileInfo, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("directory check failed for %s: %w", dir, err)
	}
	if !fileInfo.IsDir() {
		return fmt.Errorf("path %s is not a directory", dir)
	}

	tmp, err := os.CreateTemp(dir, "perm_check_")
	if err != nil {
		return fmt.Errorf("directory %s is not writable: %w", dir, err)
	}
	tmp.Close()
	os.Remove(tmp.Name())

	return nil
}
