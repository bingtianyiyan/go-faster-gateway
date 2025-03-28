package balancer

// RoundRobin will select the server in turn from the server to gateway
type RoundRobin struct {
	BaseBalancer
	i uint64
}

func init() {
	factories[R2Balancer] = NewRoundRobin
}

// NewRoundRobin create new RoundRobin balancer
func NewRoundRobin(hosts []*Node) Balancer {
	return &RoundRobin{
		i: 0,
		BaseBalancer: BaseBalancer{
			hosts: hosts,
		},
	}
}

// Balance selects a suitable host according
func (r *RoundRobin) Balance(_ string) (*Node, error) {
	r.RLock()
	defer r.RUnlock()
	if len(r.hosts) == 0 {
		return nil, NoHostError
	}
	host := r.hosts[r.i%uint64(len(r.hosts))]
	r.i++
	return host, nil
}
