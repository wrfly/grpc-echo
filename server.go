package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type server struct{}

func (s *server) Hi(ctx context.Context, x *Msg) (*Msg, error) {
	log.Printf("client send: [%s]", x.GetMsg())
	return x, nil
}

func (s *server) Sleep(ctx context.Context, x *Msg) (*Msg, error) {
	log.Printf("client sleep: [%d]", x.GetSleep())
	time.Sleep(time.Second * time.Duration(x.GetSleep()))
	return x, nil
}

func runServer(port int) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer(
		grpc.ConnectionTimeout(time.Second),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionAge:  time.Second * 5,
			MaxConnectionIdle: time.Second * 10,
			Timeout:           time.Second * 20,
		}),
		grpc.KeepaliveEnforcementPolicy(
			keepalive.EnforcementPolicy{
				// MinTime: time.Second,
				// PermitWithoutStream: true,
			}),
		grpc.MaxConcurrentStreams(5),
	)
	RegisterEchoServer(s, &server{})
	// Register reflection service on gRPC server.
	// reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
