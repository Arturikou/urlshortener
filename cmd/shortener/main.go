package main

import (
	"github.com/Arturikou/urlshortener/internal/router"
	"net/http"
)

func main() {
	r := router.New()
	err := http.ListenAndServe(`:8080`, r)
	if err != nil {
		panic(err)
	}
}
