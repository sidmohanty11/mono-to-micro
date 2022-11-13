package main

import (
	"auth-service/data"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const webPort = "80"

var counts int64

type Config struct {
	DB     *sql.DB
	Models data.Models
}

func main() {
	log.Println("-----------Starting auth service---------")

	// DB connection
	conn := connectToDB()
	if conn == nil {
		log.Println("Failed to connect to DB")
		return
	}

	app := Config{
		DB:     conn,
		Models: data.New(conn),
	}

	srv := &http.Server{
		Addr:    ":" + webPort,
		Handler: app.routes(),
	}

	err := srv.ListenAndServe()

	if err != nil {
		log.Fatal(err)
	}
}

// open db
func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func connectToDB() *sql.DB {
	dsn := os.Getenv("DSN")

	for {
		db, err := openDB(dsn)
		if err != nil {
			log.Println(err)
			counts++
		} else {
			log.Println("Connected to DB")
			return db
		}

		if counts > 10 {
			log.Println(err)
			return nil
		}

		log.Println("Backing off for two secs")
		time.Sleep(2 * time.Second)
		continue
	}
}
