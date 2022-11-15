package main

import (
	"context"
	"log"
	"logger/data"
	"net"
	"net/http"
	"net/rpc"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	webPort  = "80"
	rpcPort  = "5001"
	mongoURL = "mongodb://mongo:27017"
	gRPCPort = "50001"
)

var client *mongo.Client

type Config struct {
	Models data.Models
}

func main() {
	mongoClient, err := connectToMongo()
	if err != nil {
		log.Fatal(err)
	}
	client = mongoClient

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	app := Config{
		Models: data.New(client),
	}

	err = rpc.Register(new(RPCServer))
	go app.rpcListen()

	go app.gRPCListen()

	app.serve()
}

func (app *Config) serve() {
	srv := &http.Server{
		Addr:    ":" + webPort,
		Handler: app.routes(),
	}

	log.Println("Starting server on port", webPort)

	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func (app *Config) rpcListen() {
	log.Println("Starting RPC server on port: ", rpcPort)

	listen, err := net.Listen("tcp", "0.0.0.0:"+rpcPort)
	if err != nil {
		log.Fatal(err)
	}

	defer listen.Close()

	for {
		rpcConn, err := listen.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go rpc.ServeConn(rpcConn)
	}
}

func connectToMongo() (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(mongoURL)
	clientOptions.SetAuth(options.Credential{
		Username: "admin",
		Password: "password",
	})
	c, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Println("Error Connecting:", err.Error())
		return nil, err
	}
	return c, nil
}
