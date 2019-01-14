package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/keepalive"
)

func runClient(servers []string) {

	testLB := RegisterListLB("test", servers)
	// go func() {
	// 	time.Sleep(time.Second * 3)
	// 	testLB.UpdateBackends(servers[1:])
	// }()

	conn, err := grpc.Dial(testLB.Target(),
		grpc.WithInsecure(),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:    time.Second * 18,
			Timeout: time.Second * 17,
		}),
		grpc.WithBalancerName(roundrobin.Name),
	)

	if err != nil {
		log.Panicf("dial err: %s", err)
	}

	ctx := context.Background()

	go func() {
		for {
			state := conn.GetState()
			if conn.WaitForStateChange(ctx, state) {
				log.Printf("stage change %s->%s", state, conn.GetState())
			}
		}
	}()

	client := NewEchoClient(conn)

	// // sleep test
	// for i := 0; i < 100; i++ {
	// 	go func(index int) {
	// 		client.Sleep(ctx, &Msg{Sleep: 5})
	// 		log.Printf("[%d] sleep", index)
	// 	}(i)
	// }
	// log.Printf("---")

	var input string
	for {
		fmt.Printf("input: ")
		fmt.Scanln(&input)
		got, err := client.Hi(ctx, &Msg{Msg: input})
		if err != nil {
			log.Printf("error: %s\n", err)
			continue
		}
		if input != got.GetMsg() {
			panic(input)
		}
		input = ""
	}
}
