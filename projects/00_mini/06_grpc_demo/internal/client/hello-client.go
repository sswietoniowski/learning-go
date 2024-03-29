package client

import (
	"context"
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
