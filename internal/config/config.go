package config

import (
	"flag"
	hc "github.com/Arturikou/urlshortener/internal/handlers/config"
	"os"
)

type Config struct {
	ServerAddr string
	Handlers   hc.Config
}

func New() *Config {
	cfg := &Config{}
	flag.StringVar(&cfg.ServerAddr, "a", "localhost:8080", "The address to listen on for HTTP requests.")
	flag.StringVar(&cfg.Handlers.BaseAddr, "b", "http://localhost:8080", "The address for shortener response")
	flag.Parse()

	if envServerAddr := os.Getenv("SERVER_ADDRESS"); envServerAddr != "" {
		cfg.ServerAddr = envServerAddr
	}

	if envBaseAddr := os.Getenv("BASE_URL"); envBaseAddr != "" {
		cfg.Handlers.BaseAddr = envBaseAddr
	}

	return cfg
}
