# makefile for echo service

pb:
	protoc -I . echo.proto --go_out=plugins=grpc:.

server:
	go run . -mode s

client:
	go run . -mode c