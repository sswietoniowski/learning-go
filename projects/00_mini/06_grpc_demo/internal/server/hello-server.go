package server

import (
	"context"

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
