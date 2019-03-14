package simple

import (
	"fmt"
	"log"

	r "google.golang.org/grpc/resolver"
)

const simpleResolver = "simple"

type rBuilder struct{}

func newResolveBuilder() r.Builder {
	return &rBuilder{}
}

func (b *rBuilder) Build(target r.Target,
	cc r.ClientConn, opts r.BuildOption) (r.Resolver, error) {

	log.Printf("build resolver for cc: %p, target: %+v",
		cc, target)

	rsv := newResolver(target, cc)
	rsv.ResolveNow(r.ResolveNowOption{})
	return rsv, nil
}

func (b *rBuilder) Scheme() string {
	log.Printf("someone called scheme")
	return simpleResolver
}

func init() {
	r.Register(newResolveBuilder())
}

func Target(servers []string) string {
	t := fmt.Sprintf("%s:///%v", simpleResolver, servers)
	log.Printf("return target (%s)", t)
	return t
}
