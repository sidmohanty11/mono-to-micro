package main

import (
	"log"
	"net/http"
)

const PORT = "80"

type Config struct{}

func main() {
	app := Config{}

	log.Println("Starting server on port", PORT)

	srv := &http.Server{
		Addr:    ":" + PORT,
		Handler: app.routes(),
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
