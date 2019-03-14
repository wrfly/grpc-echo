# makefile for echo service

build: pb
	go build -o bin/grpc-echo main/*.go

.PHONY: pb
pb:
	protoc -I pb echo.proto --go_out=plugins=grpc:pb

server: build
	bin/grpc-echo -mode s

client: build
	bin/grpc-echo -mode c

# bin/grpc-echo -mode s -port 50001
# bin/grpc-echo -mode s -port 50002
# bin/grpc-echo -mode c -servers localhost:50001,localhost:50002

cs: build
	bin/grpc-echo -mode c -servers localhost:50001,localhost:50002

ss: build
	bin/grpc-echo -mode s -ports 50001,50002

s1: build
	bin/grpc-echo -mode s -ports 50001

s2: build
	bin/grpc-echo -mode s -ports 50002