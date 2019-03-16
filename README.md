# grpc-echo

This is a simple gRPC Client/Server daemon.

Use `go mod vendor` to download the dependencies.

There are several commands in Makefile:

- `make pb`: generate proto-buff codes,
    you may don't need it since I have already generated them
    unless you want to add new commands in the echo service.
- `make build`: build the package

- `make server`: run a *server* listens on localhost:5001
- `make client`: run a *client* connects to localhost:5001

- `make server-1`: run a *server* listens on localhost:5001
- `make server-2`: run a *server* listens on localhost:5002
- `make client-12`: run a *client* connect to localhost:5001 and localhost:5002

- `make servers 1201,1202,1203`: run a *server* listens on these ports
- `make clients localhost:1201,localhost:1202,localhost:1203`: run a *client* connect to these servers