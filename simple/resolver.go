package simple

import (
	"log"
	"strings"

	r "google.golang.org/grpc/resolver"
)

type resolver struct {
	servers []string
	cc      r.ClientConn
}

func (rsv *resolver) ResolveNow(opt r.ResolveNowOption) {
	log.Printf("cc(%p) resolve now with servers: %v", rsv.cc, rsv.servers)
	addrs := make([]r.Address, 0, len(rsv.servers))
	for _, e := range rsv.servers {
		addrs = append(addrs, r.Address{Addr: e})
	}
	rsv.cc.NewAddress(addrs)
}

func (r *resolver) Close() { log.Printf("cc %p closed", r.cc) }

func newResolver(target r.Target, cc r.ClientConn) r.Resolver {
	e := strings.TrimFunc(target.Endpoint,
		func(r rune) bool { return r == '[' || r == ']' })
	return &resolver{
		servers: strings.Split(e, " "),
		cc:      cc,
	}
}
