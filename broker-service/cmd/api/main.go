package main

import (
	"log"
	"net/http"
)

const webPort = "80"

type Config struct{}

func main() {
	app := Config{}

	log.Println("Starting web server on port", webPort)

	srv := &http.Server{
		Addr:    ":" + webPort,
		Handler: app.Routes(),
	}

	// start the server
	log.Fatal(srv.ListenAndServe())
}
