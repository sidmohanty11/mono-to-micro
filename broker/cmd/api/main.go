package main

import (
	"log"
	"math"
	"net/http"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const PORT = "80"

type Config struct {
	Rabbit *amqp.Connection
}

func main() {
	rabbitConn, err := connect()
	if err != nil {
		log.Fatal(err)
	}
	app := Config{
		Rabbit: rabbitConn,
	}
	defer rabbitConn.Close()
	log.Println("Starting server on port", PORT)

	srv := &http.Server{
		Addr:    ":" + PORT,
		Handler: app.routes(),
	}

	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func connect() (*amqp.Connection, error) {
	var counts int64
	var backoff = 1 * time.Second

	var connection *amqp.Connection

	for {
		c, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
		if err != nil {
			log.Println("failed to connect to rabbitmq, retrying in 1 second")
			counts++
		} else {
			log.Println("connected to rabbitmq")
			connection = c
			break
		}
		if counts > 5 {
			log.Println("failed to connect to rabbitmq")
			return nil, err
		}
		backoff = time.Duration(math.Pow(float64(counts), 2)) * time.Second
		log.Println("backing off: ", backoff)
		time.Sleep(backoff)
		continue
	}

	return connection, nil
}
