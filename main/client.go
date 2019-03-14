package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc/balancer/roundrobin"
	// "google.golang.org/grpc/keepalive"
	"google.golang.org/grpc"

	"github.com/wrfly/grpc-echo/pb"
	"github.com/wrfly/grpc-echo/simple"
)

func runClient(servers []string) {
	target := servers[0]
	if len(servers) == 2 {
		target = simple.Target(servers)
	}
	conn, err := grpc.Dial(
		target,

		// some options
		grpc.WithInsecure(),
		grpc.WithBalancerName(roundrobin.Name),

		// block until connected
		grpc.WithBlock(),

		// backoff policy
		// grpc.WithBackoffConfig(grpc.BackoffConfig{
		// 	MaxDelay: time.Second,
		// }),
		// grpc.WithBackoffMaxDelay(time.Second),

		// disable healthcheck, seems not working
		// grpc.WithDisableHealthCheck(),

		// maybe works under high corrency
		// grpc.WithDisableRetry(),

		// care of this config, read the comments carefully
		// grpc.WithKeepaliveParams(keepalive.ClientParameters{
		// 	Time:    time.Second,
		// 	Timeout: time.Second * 5,
		// }),
	)
	if err != nil {
		log.Panicf("dial err: %s", err)
	}
	defer conn.Close()
	go printStateChange(conn, "conn")

	client := pb.NewEchoClient(conn)
	log.Printf("---")
	for input := ""; ; input = "" {
		input = fmt.Sprint(time.Now().Second())
		got, err := client.Hi(context.Background(), &pb.Msg{Msg: input})
		if err != nil {
			log.Printf("error: %s", err)
			time.Sleep(time.Second * 5)
			continue
		}
		log.Printf("send: %s", got.GetMsg())
		time.Sleep(time.Second)
	}
}

func printStateChange(conn *grpc.ClientConn, name string) {
	log.Printf("conn stat: %s", conn.GetState())
	for {
		state := conn.GetState()
		if conn.WaitForStateChange(context.Background(), state) {
			log.Printf("[%s] stage change %s->%s",
				name, state, conn.GetState())
		}
	}
}
