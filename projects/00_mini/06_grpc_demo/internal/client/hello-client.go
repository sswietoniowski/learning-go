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

	resp, err := client.SayHello(ctx, &pb.SayHelloRequest{})
	if err != nil {
		log.Fatalf("Failed to call SayHello: %v", err)
	}
	
	log.Printf("Received response: %v", resp.Message)
}

func CallSayHelloServerStreaming(client pb.GreetServiceClient, names *pb.SayHelloServerStreamingRequest) {
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

func CallSayHelloClientStreaming(client pb.GreetServiceClient, names []string) {
	log.Printf("Calling SayHelloClientStreaming with names: %v", names)

	stream, err := client.SayHelloClientStreaming(context.Background())
	if err != nil {
		log.Fatalf("Failed to call SayHelloClientStreaming: %v", err)
	}

	for _, name := range names {
		duration := 2 * time.Second // simulate some processing time
		time.Sleep(duration)

		req := &pb.SayHelloClientStreamingRequest{
			Name: name,
		}
		if err := stream.Send(req); err != nil {
			log.Fatalf("Failed to send a name: %v", err)
		}

		log.Printf("Sent name: %v", name)
	}

	log.Println("Finished sending names")

	res, err := stream.CloseAndRecv()
	switch err {
	case io.EOF:
		log.Println("Received EOF")
	case nil:
		log.Printf("Received response: %v", res.Messages)
	default:
		log.Fatalf("Failed to receive a response: %v", err)
	}
}

func CallSayHelloBidirectionalStreaming(client pb.GreetServiceClient, names []string) {
	log.Printf("Calling SayHelloBidirectionalStreaming with names: %v", names)

	stream, err := client.SayHelloBidirectionalStreaming(context.Background())
	if err != nil {
		log.Fatalf("Failed to call SayHelloBidirectionalStreaming: %v", err)
	}

	go func() {
		for _, name := range names {
			duration := 2 * time.Second // simulate some processing time
			time.Sleep(duration)

			req := &pb.SayHelloBidirectionalStreamingRequest{
				Name: name,
			}
			if err := stream.Send(req); err != nil {
				log.Fatalf("Failed to send a name: %v", err)
			}

			log.Printf("Sent name: %v", name)
		}

		if err := stream.CloseSend(); err != nil {
			log.Fatalf("Failed to close the send stream: %v", err)
		}
	}()

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Failed to receive a response: %v", err)
		}

		log.Printf("Received message: %v", res.Message)
	}

	log.Println("Finished receiving messages")
}
