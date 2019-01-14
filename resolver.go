package main

import (
	"context"
	"fmt"
	"math/rand"
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

var (
	etcdRBuilder resolveBuilder
	listRBuilder resolveBuilder

	dialOptions = []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBalancerName(roundrobin.Name),
	}
)

func init() {
	etcdRBuilder = &rBuilder{
		scheme:   etcdBuilderScheme,
		rrs:      make(map[string]*rResolver, 0),
		backends: make(map[string][]string, 0),
	}

	listRBuilder = &rBuilder{
		scheme:   listBuilderScheme,
		rrs:      make(map[string]*rResolver, 0),
		backends: make(map[string][]string, 0),
	}

	resolver.Register(etcdRBuilder)
	resolver.Register(listRBuilder)
}

type Resolver interface {
	Errors() <-chan error
	Target() string
	DialWithRR(opts ...grpc.DialOption) (*grpc.ClientConn, error)
	ReConnect() error
}

// rResolver implements the resolver.Resolver and
// has an `UpdateBackends` function to update its servers
type rResolver struct {
	ccs []resolver.ClientConn

	id      string
	m       sync.Mutex
	etcdURL string
	scheme  string

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

	fmt.Printf("%s update backends %v\n", rr.id, avaliable)

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
	for _, cc := range rr.ccs {
		cc.NewAddress(addresses)
	}
}

func (rr *rResolver) ReConnect() error {
	// check avaliable endpoints
	avaliable, _ := healthCheck(rr.backends)
	if len(avaliable) == 0 {
		return fmt.Errorf("no backends avaliable")
	}
	rr.ResolveNow(resolver.ResolveNowOption{})
	return nil
}

// Close closes the resolver.
func (rr *rResolver) Close() {
	// close file or close net conn
}

func (rr *rResolver) Errors() <-chan error {
	return rr.errChan
}

func (rr *rResolver) Target() string {
	return rr.scheme + ":///" + rr.id
}

func (rr *rResolver) DialWithRR(opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(rr.Target(), dialOptions...)
	if err != nil {
		return nil, err
	}
	return conn, nil
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

func (rr *rResolver) watchBackends() {
	conn, err := rr.DialWithRR()
	if err != nil {
		panic(err) // panic at first time
	}
	for {
		state := conn.GetState()
		if conn.WaitForStateChange(context.Background(), state) {
			if conn.GetState() != state {
				rr.ResolveNow(resolver.ResolveNowOption{})
			}
		}
	}
}

func (rr *rResolver) queryKey() ([]string, error) {
	// TODO:
	return []string{"localhost:50001", "localhost:50002"}, nil
}

func (rr *rResolver) watchKey() {
	// TODO:
	// newBackends := nil
	// rb.rr.updateBackends(newBackends)
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

type resolveBuilder interface {
	Scheme() string
	AddResolver(r *rResolver)
	Build(resolver.Target, resolver.ClientConn, resolver.BuildOption) (resolver.Resolver, error)
}

// rBuilder is resolver builder
type rBuilder struct {
	scheme   string
	m        sync.Mutex
	rrs      map[string]*rResolver
	backends map[string][]string
}

// Build a resolver
func (rb *rBuilder) Build(target resolver.Target, cc resolver.ClientConn,
	opts resolver.BuildOption) (resolver.Resolver, error) {

	rb.m.Lock()
	defer rb.m.Unlock()

	id := target.Endpoint
	rr := rb.rrs[id]
	rr.ccs = append(rr.ccs, cc)
	backends := rb.backends[id]

	return rr, rr.updateBackends(backends)
}

// AddResolver can add a resolver to this builder
func (rb *rBuilder) AddResolver(r *rResolver) {
	rb.m.Lock()
	rb.rrs[r.id] = r
	rb.backends[r.id] = r.backends
	rb.m.Unlock()
}

// Scheme returns the lb scheme
func (rb *rBuilder) Scheme() string {
	return rb.scheme
}

// Register register new list builder LoadBalancer
func Register(etcdURL string, initialBackends ...string) (Resolver, error) {
	var builder resolveBuilder
	builder = listRBuilder
	if etcdURL != "" {
		builder = etcdRBuilder
	}
	rr := &rResolver{
		scheme:  builder.Scheme(),
		etcdURL: etcdURL,
		id:      randSeq(9),
		ccs:     make([]resolver.ClientConn, 0),
		errChan: make(chan error, 10),
	}

	// etcd get initialBackends
	if rr.scheme == etcdBuilderScheme {
		remoteBackends, err := rr.queryKey()
		if err != nil {
			return nil, err
		}
		initialBackends = remoteBackends
		go rr.watchKey()
	}
	rr.backends = initialBackends

	if rr.scheme == etcdBuilderScheme {
		etcdRBuilder.AddResolver(rr)
	} else {
		listRBuilder.AddResolver(rr)
	}

	go rr.watchBackends()

	return rr, nil
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
