package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"github.com/wrfly/grpc-echo/pb"
)

type server struct {
	port string
}

func (s *server) Hi(ctx context.Context, x *pb.Msg) (*pb.Msg, error) {
	log.Printf("[%s] got: [%s]", s.port, x.GetMsg())
	return x, nil
}

func (s *server) Sleep(ctx context.Context, x *pb.Msg) (*pb.Msg, error) {
	log.Printf("client sleep: [%d]", x.GetSleep())
	time.Sleep(time.Second * time.Duration(x.GetSleep()))
	return x, nil
}

func runServer(port string) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer(
		grpc.ConnectionTimeout(time.Second),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: time.Second * 10,
			Timeout:           time.Second * 20,
		}),
		grpc.KeepaliveEnforcementPolicy(
			keepalive.EnforcementPolicy{
				MinTime:             time.Second,
				PermitWithoutStream: true,
			}),
		grpc.MaxConcurrentStreams(5),
	)
	pb.RegisterEchoServer(s, &server{port})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
