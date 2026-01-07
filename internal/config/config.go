package config

import (
	"flag"
	hc "github.com/Arturikou/urlshortener/internal/handlers/config"
)

type Config struct {
	ServerAddr string
	Handlers   hc.Config
}

func New() *Config {
	cfg := &Config{}
	flag.StringVar(&cfg.ServerAddr, "server-addr", "localhost:8080", "The address to listen on for HTTP requests.")
	flag.StringVar(&cfg.Handlers.BaseAddr, "base-addr", "http://localhost:8080", "The address for url response")

	return cfg
}
