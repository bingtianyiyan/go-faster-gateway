package balancer

import (
	"errors"
)

var (
	NoHostError                = errors.New("no host")
	AlgorithmNotSupportedError = errors.New("algorithm not supported")
)

// Balancer interface is the load balancer for the reverse gateway
type Balancer interface {
	Add(node *Node)
	Remove(string)
	Balance(string) (*Node, error)
	Inc(string)
	Done(string)
}

// Factory is the factory that generates Balancer,
// and the factory design pattern is used here
type Factory func([]*Node) Balancer

var factories = make(map[string]Factory)

// Build generates the corresponding Balancer according to the algorithm
func Build(algorithm string, hosts []*Node) (Balancer, error) {
	factory, ok := factories[algorithm]
	if !ok {
		return nil, AlgorithmNotSupportedError
	}
	return factory(hosts), nil
}
