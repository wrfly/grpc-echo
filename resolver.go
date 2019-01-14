package main

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
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
	DialWithRR(opts ...grpc.DialOption) (*grpc.ClientConn, error)
	Errors() <-chan error
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
	errChan  chan error
}

// ResolveNow could be called multiple times concurrently.
func (rr *rResolver) ResolveNow(opts resolver.ResolveNowOption) {
	rr.m.Lock()
	defer rr.m.Unlock()

	// check avaliable endpoints
	avaliable, failed := healthCheck(rr.backends)
	if len(failed) != 0 {
		err := fmt.Errorf("failed backends: %v", failed)
		select {
		case rr.errChan <- err:
		default:
		}
	}

	hash := fmt.Sprint(avaliable)
	if rr.hash == hash {
		return
	}
	rr.hash = hash

	fmt.Printf("update backends %v\n", avaliable)

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

func (rr *rResolver) updateBackends(backends []string) error {
	// health check first
	aliveBackends, failedBackends := healthCheck(backends)

	// update backends
	rr.backends = aliveBackends

	// resolve immediately
	rr.ResolveNow(resolver.ResolveNowOption{})

	if len(failedBackends) != 0 {
		err := fmt.Errorf("failed backends: %v", failedBackends)
		select {
		case rr.errChan <- err:
			return err
		default:
			return err
		}
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

	rb.rr.cc = cc
	globalListMutex.Lock()
	backends := globalListEndpoints[target.Endpoint]
	globalListMutex.Unlock()

	return rb.rr, rb.rr.updateBackends(backends)
}

// Scheme returns the lb scheme
func (rb *rBuilder) Scheme() string {
	return rb.scheme
}

func (rb *rBuilder) Target() string {
	return rb.scheme + ":///" + rb.id
}

func (rb *rBuilder) DialWithRR(opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return grpc.Dial(rb.Target(),
		grpc.WithInsecure(),
		grpc.WithBalancerName(roundrobin.Name),
	)
}

func (rb *rBuilder) Errors() <-chan error {
	return rb.rr.errChan
}

func (rb *rBuilder) watchBackends() {
	conn, err := rb.DialWithRR()
	if err != nil {
		panic(err) // panic at first time
	}
	for {
		state := conn.GetState()
		if conn.WaitForStateChange(context.Background(), state) {
			if conn.GetState() != state {
				rb.resolve()
			}
		}
	}
}

func (rb *rBuilder) resolve() {
	rb.rr.ResolveNow(resolver.ResolveNowOption{})
}

func (rb *rBuilder) queryKey() ([]string, error) {
	// TODO:
	return []string{"localhost:50001", "localhost:50002"}, nil
}

func (rb *rBuilder) watchKey() {
	// TODO:
	// newBackends := nil
	// rb.rr.updateBackends(newBackends)
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
		rr: &rResolver{
			errChan: make(chan error, 10),
		},
	}

	// etcd get initialBackends
	if rb.scheme == etcdBuilderScheme {
		gotBackends, err := rb.queryKey()
		if err != nil {
			return nil, err
		}
		go rb.watchKey()
		initialBackends = gotBackends
	}
	resolver.Register(rb)

	globalListMutex.Lock()
	globalListEndpoints[rb.id] = initialBackends
	globalListMutex.Unlock()

	go rb.watchBackends()

	return rb, nil
}
