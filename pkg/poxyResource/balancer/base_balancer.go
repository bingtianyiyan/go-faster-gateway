package balancer

import (
	"sync"
)

type BaseBalancer struct {
	sync.RWMutex
	hosts []*Node
}

// Add new host to the balancer
func (b *BaseBalancer) Add(host *Node) {
	b.Lock()
	defer b.Unlock()
	for _, h := range b.hosts {
		if h.Service == host.Service {
			return
		}
	}
	b.hosts = append(b.hosts, host)
}

// Remove new host from the balancer
func (b *BaseBalancer) Remove(host string) {
	b.Lock()
	defer b.Unlock()
	for i, h := range b.hosts {
		if h.Service == host {
			b.hosts = append(b.hosts[:i], b.hosts[i+1:]...)
			return
		}
	}
}

// Balance selects a suitable host according
func (b *BaseBalancer) Balance(key string) (string, error) {
	return "", nil
}

// Inc .
func (b *BaseBalancer) Inc(_ string) {}

// Done .
func (b *BaseBalancer) Done(_ string) {}
