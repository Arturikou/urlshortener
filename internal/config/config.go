package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type HandlersConfig struct {
	BaseAddr string
}

type DatabaseConfig struct {
	DSN string
}

type Audit struct {
	AuditFile string
	AuditURL  string
}

type Config struct {
	ServerAddr      string
	LogLevel        string
	FileStoragePath string
	Handlers        HandlersConfig
	Database        DatabaseConfig
	Audit           Audit
}

func New() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.ServerAddr, "a", "localhost:8080", "The address to listen on for HTTP requests.")
	flag.StringVar(&cfg.LogLevel, "level", "info", "Log level")
	flag.StringVar(&cfg.FileStoragePath, "f", "storage.json", "File storage path")
	flag.StringVar(&cfg.Handlers.BaseAddr, "b", "http://localhost:8080", "The address for shortener response")
	flag.StringVar(&cfg.Database.DSN, "d", "", "Database DSN")
	flag.StringVar(&cfg.Audit.AuditFile, "httpaudit-file", "", "Audit file path")
	flag.StringVar(&cfg.Audit.AuditURL, "httpaudit-url", "", "Audit URL")
	flag.Parse()

	if envServerAddr := os.Getenv("SERVER_ADDRESS"); envServerAddr != "" {
		cfg.ServerAddr = envServerAddr
	}

	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		cfg.LogLevel = envLogLevel
	}

	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		cfg.FileStoragePath = envFileStoragePath
	}

	if envBaseAddr := os.Getenv("BASE_URL"); envBaseAddr != "" {
		cfg.Handlers.BaseAddr = envBaseAddr
	}

	if envDatabaseDSN := os.Getenv("DATABASE_DSN"); envDatabaseDSN != "" {
		cfg.Database.DSN = envDatabaseDSN
	}

	if envAuditFilePath := os.Getenv("AUDIT_FILE"); envAuditFilePath != "" {
		cfg.Audit.AuditFile = envAuditFilePath
	}

	if envAuditURL := os.Getenv("AUDIT_URL"); envAuditURL != "" {
		cfg.Audit.AuditURL = envAuditURL
	}

	if err := validateFilePath(cfg.FileStoragePath); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	if cfg.Audit.AuditFile != "" {
		err := validateFilePath(cfg.Audit.AuditFile)
		if err != nil {
			log.Fatalf("Configuration error: %v", err)
		}
	}

	return cfg
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
