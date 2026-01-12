package main

import (
	"github.com/Arturikou/urlshortener/internal/config"
	"github.com/Arturikou/urlshortener/internal/handlers"
	"github.com/Arturikou/urlshortener/internal/repository"
	"github.com/Arturikou/urlshortener/internal/router"
	"github.com/Arturikou/urlshortener/internal/service/url"
	"log"
	"net/http"
)

func main() {
	cfg := config.New()
	storage := repository.NewStore()
	urlService := url.New(storage)
	h := handlers.New(urlService, cfg.Handlers)
	r := router.New(h)

	srv := &http.Server{
		Addr:    cfg.ServerAddr,
		Handler: r,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server stopped with error: %v", err)
	}
}
