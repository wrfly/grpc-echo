# makefile for echo service

pb:
	protoc -I . echo.proto --go_out=plugins=grpc:.

server:
	go run *.go -mode s

client:
	go run *.go -mode c

# go run *.go -mode s -port 50001
# go run *.go -mode s -port 50002
# go run *.go -mode c -servers localhost:50001,localhost:50002

cs:
	go run *.go -mode c -servers localhost:50001,localhost:50002

ss:
	go run *.go -mode s -ports 50001,50002

s1:
	go run *.go -mode s -ports 50001

s2:
	go run *.go -mode s -ports 50002