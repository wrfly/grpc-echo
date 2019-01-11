package main

import (
	"fmt"
	"log"
	"strings"

	"google.golang.org/grpc/resolver"
)

const listBuildScheme = "list"

type listResolver struct {
	target resolver.Target
	cc     resolver.ClientConn

	serverList []string
	hash       string
}

// It could be called multiple times concurrently.
func (lr *listResolver) ResolveNow(opts resolver.ResolveNowOption) {
	if lr.hash == fmt.Sprint(lr.serverList) {
		return
	}

	addresses := []resolver.Address{}
	for _, endpoint := range strings.Split(lr.target.Endpoint, ",") {
		addresses = append(addresses,
			resolver.Address{
				Addr:       endpoint,
				Type:       resolver.Backend,
				ServerName: endpoint,
			},
		)
	}
	lr.cc.NewAddress(addresses)
	lr.hash = fmt.Sprint(lr.serverList)
}

// Close closes the resolver.
func (lr *listResolver) Close() {
	// file close or net close
}

func (lr *listResolver) UpdateServerList(list []string) {
	lr.serverList = list

	lr.target.Endpoint = strings.Join(lr.serverList, ",")
	lr.ResolveNow(resolver.ResolveNowOption{})
}

type listBuilder struct{}

func (lb *listBuilder) Build(target resolver.Target, cc resolver.ClientConn,
	opts resolver.BuildOption) (resolver.Resolver, error) {
	log.Printf("build target: %+v", target)

	lr := &listResolver{
		target:     target,
		cc:         cc,
		serverList: strings.Split(target.Endpoint, ","),
	}
	lr.ResolveNow(resolver.ResolveNowOption{})

	return lr, nil
}
func (lb *listBuilder) Scheme() string {
	return listBuildScheme
}

func init() {
	resolver.SetDefaultScheme(listBuildScheme)
	resolver.Register(&listBuilder{})
}
