package main

import (
	"fmt"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc/resolver"
)

// schemes
const (
	ListBuilderScheme = "list_builder"
)

// ListBuilder can build a resolver from a backend list
// and can also update the backends
type ListBuilder interface {
	Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOption) (resolver.Resolver, error)
	UpdateBackends(backends []string) error
	Target() string
}

// ListResolver implements the resolver.Resolver
type ListResolver interface {
	ResolveNow(opts resolver.ResolveNowOption)
	Close()
}

// global list endpoints
var (
	globalListEndpoints map[string][]string
	globalListMutex     sync.Mutex
)

func init() {
	globalListMutex.Lock()
	defer globalListMutex.Unlock()
	globalListEndpoints = make(map[string][]string, 0)
	resolver.SetDefaultScheme(ListBuilderScheme)
}

// listResolver implements the resolver.Resolver and
// has an `UpdateBackends` function to update its servers
type listResolver struct {
	cc resolver.ClientConn

	m        sync.Mutex
	backends []string
	hash     string
}

// ResolveNow could be called multiple times concurrently.
func (lr *listResolver) ResolveNow(opts resolver.ResolveNowOption) {
	// update backends
	lr.m.Lock()
	avaliable, _ := healthCheck(lr.backends)
	lr.m.Unlock()

	hash := fmt.Sprint(avaliable)
	if lr.hash == hash {
		return
	}
	lr.hash = hash

	addresses := []resolver.Address{}
	for _, e := range avaliable {
		addresses = append(addresses,
			resolver.Address{
				Addr:       e,
				Type:       resolver.Backend,
				ServerName: e,
			},
		)
	}
	lr.cc.NewAddress(addresses)
}

// Close closes the resolver.
func (lr *listResolver) Close() {
	// close file or close net conn
}

func (lr *listResolver) updateBackends(list []string) error {
	// health check first
	aliveBackends, failedBackends := healthCheck(list)

	// update backends
	lr.m.Lock()
	lr.backends = aliveBackends
	lr.m.Unlock()

	fmt.Printf("updateBackends: %v %v %v %v\n", list, aliveBackends, failedBackends, lr.backends)

	// resolve immediately
	lr.ResolveNow(resolver.ResolveNowOption{})

	if len(failedBackends) != 0 {
		return fmt.Errorf("failed backends: %v", failedBackends)
	}

	return nil
}

func healthCheck(list []string) ([]string, []string) {
	aliveBackends, failedBackends := []string{}, []string{}
	for _, e := range list {
		conn, err := net.Dial("tcp", e)
		if err != nil {
			failedBackends = append(failedBackends, e)
			continue
		}
		conn.Close()
		aliveBackends = append(aliveBackends, e)
	}

	return aliveBackends, failedBackends
}

// listBuilder can build a resolver who can update backends
// from a server list
type listBuilder struct {
	name string
	lr   *listResolver
}

// Build a resolver
func (lb *listBuilder) Build(target resolver.Target, cc resolver.ClientConn,
	opts resolver.BuildOption) (resolver.Resolver, error) {
	backends := globalListEndpoints[target.Endpoint]
	lb.lr = &listResolver{
		cc:       cc,
		backends: backends,
	}
	if err := lb.lr.updateBackends(backends); err != nil {
		return nil, err
	}

	go func() {
		for {
			time.Sleep(time.Second)
			lb.lr.ResolveNow(resolver.ResolveNowOption{})
		}
	}()

	return lb.lr, nil
}

// Scheme returns the lb scheme
func (lb *listBuilder) Scheme() string {
	return ListBuilderScheme
}

func (lb *listBuilder) Target() string {
	return lb.name
}

// UpdateBackends can update the gRPC backend list
func (lb *listBuilder) UpdateBackends(backends []string) error {
	if lb.lr == nil {
		return fmt.Errorf("resolver is nil")
	}
	globalListMutex.Lock()
	defer globalListMutex.Unlock()
	globalListEndpoints[lb.name] = backends

	return lb.lr.updateBackends(backends)
}

// RegisterListLB register new list builder LoadBalancer
func RegisterListLB(name string, initialBackends []string) ListBuilder {
	lb := &listBuilder{
		name: name,
	}

	globalListMutex.Lock()
	globalListEndpoints[lb.name] = initialBackends
	globalListMutex.Unlock()

	resolver.Register(lb)
	return lb
}
