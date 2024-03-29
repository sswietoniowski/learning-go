package client

import (
	"context"
	"io"
	"log"
	"time"

	pb "github.com/sswietoniowski/learning-go/projects/00_mini/06_grpc_demo/internal/common/genproto"
)

func CallSayHello(client pb.GreetServiceClient) {
	timeout := time.Duration(5) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := client.SayHello(ctx, &pb.NoParam{})
	if err != nil {
		log.Fatalf("Failed to call SayHello: %v", err)
	}
	log.Printf("Received response: %v", resp.Message)
}

func CallSayHelloServerStreaming(client pb.GreetServiceClient, names *pb.NamesList) {
	log.Printf("Calling SayHelloServerStreaming with names: %v", names.Names)

	stream, err := client.SayHelloServerStreaming(context.Background(), names)
	if err != nil {
		log.Fatalf("Failed to call SayHelloServerStreaming: %v", err)
	}

	for {
		message, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Failed to receive a message: %v", err)
		}

		log.Printf("Received message: %v", message.Message)
	}

	log.Println("Finished receiving messages")
}
