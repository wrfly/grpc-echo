# makefile for echo service

export GO111MODULE=on

.PHONY: vendor
vendor:
	go mod vendor

.PHONY: pb
pb:
	protoc -I pb echo.proto --go_out=plugins=grpc:pb

build:
	@go build -o bin/grpc-echo main/*.go

server: build
	bin/grpc-echo -mode s

client: build
	bin/grpc-echo -mode c

server-1: build
	bin/grpc-echo -mode s -ports 5001

server-2: build
	bin/grpc-echo -mode s -ports 5002

client-12: build
	bin/grpc-echo -mode c -servers localhost:5001,localhost:5002

# with arguments
runargs := $(word 2, $(MAKECMDGOALS) )

clients: build
	bin/grpc-echo -mode c -servers $(runargs)

servers: build
	bin/grpc-echo -mode s -ports $(runargs)