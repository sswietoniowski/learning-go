package main

import (
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/sswietoniowski/learning-go/projects/00_mini/06_grpc_demo/internal/client"
	pb "github.com/sswietoniowski/learning-go/projects/00_mini/06_grpc_demo/internal/common/genproto"
)

const (
	port = ":8080"
)

func main() {
	conn, err := grpc.Dial("localhost"+port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to dial server: %v", err)
	}
	defer conn.Close()

	c := pb.NewGreetServiceClient(conn)

	client.CallSayHello(c)

	names := &pb.SayHelloServerStreamingRequest{
		Names: []string{"Alice", "Bob", "Charlie"},
	}

	client.CallSayHelloServerStreaming(c, names)

	client.CallSayHelloClientStreaming(c, []string{"Alice", "Bob", "Charlie"})

	client.CallSayHelloBidirectionalStreaming(c, []string{"Alice", "Bob", "Charlie"})
}
