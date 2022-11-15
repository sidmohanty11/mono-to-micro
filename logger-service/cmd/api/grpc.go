package main

import (
	"context"
	"log"
	"logger/data"
	"logger/logs"
	"net"

	"google.golang.org/grpc"
)

type LogServer struct {
	logs.UnimplementedLogServiceServer
	Models data.Models
}

func (l *LogServer) WriteLog(ctx context.Context, req *logs.LogRequest) (*logs.LogResponse, error) {
	input := req.GetLogEntry()

	logEntry := data.LogEntry{
		Name: input.Name,
		Data: input.Data,
	}

	err := l.Models.LogEntry.Insert(logEntry)

	if err != nil {
		res := &logs.LogResponse{Result: "failed"}
		return res, err
	}

	return &logs.LogResponse{Result: "successfully logged"}, nil
}

func (app *Config) gRPCListen() {
	lis, err := net.Listen("tcp", "0.0.0.0:"+gRPCPort)
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()

	logs.RegisterLogServiceServer(s, &LogServer{Models: app.Models})

	log.Println("Starting gRPC server on port: ", gRPCPort)
	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
