package main

import (
	"fmt"
	"net"
	"strings"
	"sync"

	"google.golang.org/grpc/resolver"
)

// ListBuilderScheme for server backend list builder
const ListBuilderScheme = "list_builder"

func init() {
	resolver.SetDefaultScheme(ListBuilderScheme)
	resolver.Register(&ListBuilder{})
}

// ListResolver implements the resolver.Resolver and
// has an `UpdateBackends` function to update its servers
type ListResolver struct {
	cc resolver.ClientConn

	m        sync.Mutex
	backends []string
	hash     string
}

// ResolveNow could be called multiple times concurrently.
func (lr *ListResolver) ResolveNow(opts resolver.ResolveNowOption) {
	lr.m.Lock()
	defer lr.m.Unlock()

	if lr.hash == fmt.Sprint(lr.backends) {
		return
	}
	lr.hash = fmt.Sprint(lr.backends)

	addresses := []resolver.Address{}
	for _, endpoint := range lr.backends {
		addresses = append(addresses,
			resolver.Address{
				Addr:       endpoint,
				Type:       resolver.Backend,
				ServerName: endpoint,
			},
		)
	}
	lr.cc.NewAddress(addresses)
}

// Close closes the resolver.
func (lr *ListResolver) Close() {
	// close file or close net conn
}

// UpdateBackends can update the gRPC backend list
func (lr *ListResolver) UpdateBackends(list []string) error {
	// preflight
	for _, e := range list {
		c, err := net.Dial("tcp", e)
		if err != nil {
			return err
		}
		c.Close()
	}

	// update backends
	lr.m.Lock()
	lr.backends = list
	lr.m.Unlock()

	// resolve immediately
	lr.ResolveNow(resolver.ResolveNowOption{})

	return nil
}

// ListBuilder can build a resolver who can update backends
// from a server list
type ListBuilder struct {
	lr *ListResolver
}

// Build a resolver
func (lb *ListBuilder) Build(target resolver.Target, cc resolver.ClientConn,
	opts resolver.BuildOption) (resolver.Resolver, error) {

	lb.lr = &ListResolver{cc: cc}
	return lb.lr, lb.lr.UpdateBackends(strings.Split(target.Endpoint, ","))
}

// Scheme returns the lb scheme
func (lb *ListBuilder) Scheme() string {
	return ListBuilderScheme
}

// UpdateBackends can update the gRPC backend list
func (lb *ListBuilder) UpdateBackends(list []string) error {
	if lb.lr == nil {
		return fmt.Errorf("resolver is nil")
	}
	return lb.lr.UpdateBackends(list)
}

// Register new list builder LoadBalancer
func Register() *ListBuilder {
	lb := &ListBuilder{}
	resolver.Register(lb)
	return lb
}
