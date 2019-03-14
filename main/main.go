package main

import (
	"flag"
	"fmt"
	"strings"
)

var (
	ports   string
	mode    string
	servers string
)

func init() {
	flag.StringVar(&ports, "ports", "50001", "server ports")
	flag.StringVar(&mode, "mode", "s", "run mode, s=server || c=client")
	flag.StringVar(&servers, "servers", "localhost:50001", "server list, split with comma")
	flag.Parse()
}

func main() {
	if mode == "s" {
		fmt.Println("run server")
		for _, port := range strings.Split(ports, ",") {
			go runServer(port)
		}
	} else if mode == "c" {
		fmt.Println("run client")
		runClient(strings.Split(servers, ","))
	} else {
		panic("unknown mode, s=server or c=client")
	}

	<-make(chan int)
}
