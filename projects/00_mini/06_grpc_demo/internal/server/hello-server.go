package server

import (
	"context"
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
