package main

import (
	"flag"
	"github.com/Arturikou/urlshortener/internal/config"
	"github.com/Arturikou/urlshortener/internal/handlers"
	"github.com/Arturikou/urlshortener/internal/repository"
	"github.com/Arturikou/urlshortener/internal/router"
	"github.com/Arturikou/urlshortener/internal/service/url"
	"net/http"
)

func main() {
	cfg := config.New()
	flag.Parse()

	storage := repository.NewStore()
	urlService := url.New(storage)
	h := handlers.New(urlService, cfg.Handlers)
	r := router.New(h)

	srv := &http.Server{
		Addr:    cfg.ServerAddr,
		Handler: r,
	}

	err := srv.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
