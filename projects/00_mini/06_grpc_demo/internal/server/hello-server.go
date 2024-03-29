package server

import pb "github.com/sswietoniowski/learning-go/projects/00_mini/06_grpc_demo/internal/common/genproto"

type HelloServer struct {
	pb.GreetServiceServer
}
