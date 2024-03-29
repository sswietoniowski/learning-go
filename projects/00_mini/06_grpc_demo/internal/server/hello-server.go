package server

import (
	"context"
	"io"
	"log"
	"time"

	pb "github.com/sswietoniowski/learning-go/projects/00_mini/06_grpc_demo/internal/common/genproto"
)

type HelloServer struct {
	pb.GreetServiceServer
}

func (s *HelloServer) SayHello(ctx context.Context, _ *pb.NoParam) (*pb.HelloResponse, error) {
	return &pb.HelloResponse{
		Message: "Hello from the gRPC server",
	}, nil
}

func (s *HelloServer) SayHelloServerStreaming(req *pb.NamesList, stream pb.GreetService_SayHelloServerStreamingServer) error {
	log.Printf("Received request with names: %v", req.Names)
	for _, name := range req.Names {
		duration := 2 * time.Second // simulate some processing time
		time.Sleep(duration)

		res := &pb.HelloResponse{
			Message: "Hello " + name,
		}
		if err := stream.Send(res); err != nil {
			return err
		}
	}

	return nil
}

func (s *HelloServer) SayHelloClientStreaming(stream pb.GreetService_SayHelloClientStreamingServer) error {
	log.Printf("Receiving names from the client...")

	names := []string{}

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		log.Printf("Received name: %v", req.Name)
		names = append(names, req.Name)
	}
	log.Printf("Received names: %v", names)

	return nil
}

func (s *HelloServer) SayHelloBidirectionalStreaming(stream pb.GreetService_SayHelloBidirectionalStreamingServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		log.Printf("Received name: %v", req.Name)

		res := &pb.HelloResponse{
			Message: "Hello " + req.Name,
		}
		if err := stream.Send(res); err != nil {
			return err
		}
	}

	return nil
}
