package config

import (
	"flag"
	hc "github.com/Arturikou/urlshortener/internal/handlers/config"
	"os"
)

type Config struct {
	ServerAddr      string
	LogLevel        string
	FileStoragePath string
	Handlers        hc.Config
}

func New() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.ServerAddr, "a", "localhost:8080", "The address to listen on for HTTP requests.")
	flag.StringVar(&cfg.LogLevel, "level", "info", "Log level")
	flag.StringVar(&cfg.FileStoragePath, "f", "storage.json", "File storage path")
	flag.StringVar(&cfg.Handlers.BaseAddr, "b", "http://localhost:8080", "The address for shortener response")
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

	return cfg
}
