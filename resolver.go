package main

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
)

// schemes
const (
	listBuilderScheme = "list"
	etcdBuilderScheme = "etcd"
)

// Builder can build a resolver from a backend list
type Builder interface {
	Target() string
}

// global list endpoints
var (
	globalListEndpoints map[string][]string
	globalListMutex     sync.Mutex
)

func init() {
	globalListEndpoints = make(map[string][]string, 0)
}

// rResolver implements the resolver.Resolver and
// has an `UpdateBackends` function to update its servers
type rResolver struct {
	cc resolver.ClientConn

	m        sync.Mutex
	backends []string
	hash     string
}

// ResolveNow could be called multiple times concurrently.
func (rr *rResolver) ResolveNow(opts resolver.ResolveNowOption) {
	// check avaliable endpoints
	avaliable, _ := healthCheck(rr.backends)

	hash := fmt.Sprint(avaliable)
	if rr.hash == hash {
		return
	}
	rr.m.Lock()
	rr.hash = hash
	rr.m.Unlock()

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
	rr.cc.NewAddress(addresses)
}

// Close closes the resolver.
func (rr *rResolver) Close() {
	// close file or close net conn
}

func (rr *rResolver) updateBackends(list []string) error {
	// health check first
	aliveBackends, failedBackends := healthCheck(list)

	// update backends
	rr.m.Lock()
	rr.backends = aliveBackends
	rr.m.Unlock()

	// resolve immediately
	rr.ResolveNow(resolver.ResolveNowOption{})

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

// rBuilder is resolver builder
type rBuilder struct {
	id      string
	rr      *rResolver
	scheme  string
	etcdURL string
}

// Build a resolver
func (rb *rBuilder) Build(target resolver.Target, cc resolver.ClientConn,
	opts resolver.BuildOption) (resolver.Resolver, error) {

	backends := globalListEndpoints[target.Endpoint]
	rb.rr = &rResolver{
		cc:       cc,
		backends: backends,
	}
	if err := rb.rr.updateBackends(backends); err != nil {
		return nil, err
	}

	return rb.rr, nil
}

// Scheme returns the lb scheme
func (rb *rBuilder) Scheme() string {
	return rb.scheme
}

func (rb *rBuilder) Target() string {
	return rb.scheme + ":///" + rb.id
}

func (rb *rBuilder) updateBackends(backends []string) error {
	if rb.rr == nil {
		return fmt.Errorf("resolver is nil")
	}
	globalListMutex.Lock()
	defer globalListMutex.Unlock()
	globalListEndpoints[rb.id] = backends

	return rb.rr.updateBackends(backends)
}

func (rb *rBuilder) watch() {
	conn, err := grpc.Dial(rb.Target(), grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	if err == nil {
		for {
			state := conn.GetState()
			if conn.WaitForStateChange(ctx, state) {
				fmt.Println("state change, resolve now")
				rb.resolve()
			}
		}
	}
}

func (rb *rBuilder) resolve() {
	rb.rr.ResolveNow(resolver.ResolveNowOption{})
}

func (rb *rBuilder) queryKey(etcd string) ([]string, error) {
	// TODO:
	return []string{"localhost:50001", "localhost:50002"}, nil
}

func (rb *rBuilder) watchKey(etcd string) []string {
	// TODO:
	return nil
}

// Register register new list builder LoadBalancer
func Register(etcdURL string, initialBackends ...string) (Builder, error) {
	scheme := listBuilderScheme
	if etcdURL != "" {
		scheme = etcdBuilderScheme
	}
	rb := &rBuilder{
		id:      fmt.Sprint(time.Now().Unix()),
		scheme:  scheme,
		etcdURL: etcdURL,
	}

	// etcd get initialBackends
	if rb.scheme == etcdBuilderScheme {
		gotBackends, err := rb.queryKey(etcdURL)
		if err != nil {
			return nil, err
		}
		go rb.watchKey(etcdURL)
		initialBackends = gotBackends
	}

	globalListMutex.Lock()
	globalListEndpoints[rb.id] = initialBackends
	globalListMutex.Unlock()

	resolver.Register(rb)

	go rb.watch()

	return rb, nil
}
