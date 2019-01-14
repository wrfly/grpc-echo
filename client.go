package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
)

func runClient(servers []string) {

	testRR, err := Register("/etcd/key")
	if err != nil {
		log.Panicf("register error: %s", err)
	}
	go func() {
		for err := range testRR.Errors() {
			log.Printf("resolver error: %s", err)
		}
	}()

	conn, err := testRR.DialWithRR()
	if err != nil {
		log.Panicf("dial err: %s", err)
	}
	log.Printf("conn stat: %s", conn.GetState())
	defer conn.Close()
	go printStateChange(conn, "conn")

	conn2, err := testRR.DialWithRR()
	if err != nil {
		log.Panicf("dial err: %s", err)
	}
	log.Printf("conn2 stat: %s", conn2.GetState())
	defer conn2.Close()
	go printStateChange(conn2, "conn2")

	client := NewEchoClient(conn)

	// sleep test
	// for i := 0; i < 100; i++ {
	// 	go func(index int) {
	// 		client.Sleep(ctx, &Msg{Sleep: 5})
	// 		log.Printf("[%d] sleep", index)
	// 	}(i)
	// }
	// log.Printf("---")

	for input := ""; ; input = "" {
		input = fmt.Sprint(time.Now().Second())
		got, err := client.Hi(context.Background(), &Msg{Msg: input})
		if err != nil {
			log.Printf("error: %s\n", err)
			time.Sleep(time.Second * 5)
			if err := testRR.ReConnect(); err != nil {
				log.Printf("reconnect error: %s\n", err)
			}
			continue
		}
		if input != got.GetMsg() {
			panic(input)
		}
		time.Sleep(time.Second)
	}
}

func printStateChange(conn *grpc.ClientConn, name string) {
	for {
		state := conn.GetState()
		if conn.WaitForStateChange(context.Background(), state) {
			log.Printf("[%s] stage change %s->%s",
				name, state, conn.GetState())
		}
	}
}
