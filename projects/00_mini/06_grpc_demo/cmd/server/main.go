package main

import (
	"log"
	"net"

	pb "github.com/sswietoniowski/learning-go/projects/00_mini/06_grpc_demo/internal/common/genproto"
	"github.com/sswietoniowski/learning-go/projects/00_mini/06_grpc_demo/internal/server"
	"google.golang.org/grpc"
)

const (
	port = ":8080"
)

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	srv := server.HelloServer{}
	pb.RegisterGreetServiceServer(grpcServer, &srv)

	log.Printf("Server started at %v", lis.Addr())

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
