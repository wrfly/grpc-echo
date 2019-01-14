package main

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
)

func runClient(servers []string) {

	testLB, err := Register("/etcd/key")
	if err != nil {
		log.Panicf("register error: %s", err)
	}

	go func() {
		for err := range testLB.Errors() {
			log.Printf("lb error: %s", err)
		}
	}()

	conn, err := grpc.Dial(testLB.Target(),
		grpc.WithInsecure(),
		grpc.WithBalancerName(roundrobin.Name),
	)
	if err != nil {
		log.Panicf("dial err: %s", err)
	}

	// conn, err := testLB.DialWithRR(
	// 	grpc.WithKeepaliveParams(keepalive.ClientParameters{
	// 		Time:    time.Second * 18,
	// 		Timeout: time.Second * 17,
	// 	}),
	// )
	// if err != nil {
	// 	log.Panicf("dial err: %s", err)
	// }

	// conn2, err := testLB.DialWithRR(
	// 	grpc.WithKeepaliveParams(keepalive.ClientParameters{
	// 		Time:    time.Second * 18,
	// 		Timeout: time.Second * 17,
	// 	}),
	// )
	// if err != nil {
	// 	log.Panicf("dial err: %s", err)
	// }
	// conn2.Close()

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
