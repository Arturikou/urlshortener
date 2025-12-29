package main

import (
	"github.com/Arturikou/urlshortener/internal/handlers"
	"github.com/Arturikou/urlshortener/internal/repository"
	"github.com/Arturikou/urlshortener/internal/router"
	"github.com/Arturikou/urlshortener/internal/service"
	"net/http"
)

func main() {
	storage := repository.NewStore()
	urlsService := service.New(storage)
	h := handlers.New(urlsService)
	r := router.New(h)

	err := http.ListenAndServe(`:8080`, r)
	if err != nil {
		panic(err)
	}
}
