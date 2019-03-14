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

	addrs := []r.Address{}
	for _, e := range rsv.servers {
		addrs = append(addrs,
			r.Address{
				Addr:       e,
				Type:       r.Backend,
				ServerName: e,
			},
		)
	}
	rsv.cc.NewAddress(addrs)
}

func (r *resolver) Close() {
	log.Printf("cc %p closed", r.cc)
}

func newResolver(target r.Target, cc r.ClientConn) r.Resolver {
	e := target.Endpoint
	e = strings.TrimPrefix(e, "[")
	e = strings.TrimSuffix(e, "]")

	servers := strings.Split(e, " ")
	log.Printf("new resolver for cc %p, servers: %v", cc, servers)
	return &resolver{
		servers: servers,
		cc:      cc,
	}
}
