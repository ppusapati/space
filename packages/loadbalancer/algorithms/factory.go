package algorithms

import (
	"fmt"

	"p9e.in/samavaya/packages/loadbalancer"
)

// New creates a new load balancer based on the algorithm type
func New(algorithm loadbalancer.Algorithm, opts ...loadbalancer.Option) (loadbalancer.LoadBalancer, error) {
	options := loadbalancer.DefaultOptions()
	for _, opt := range opts {
		opt(&options)
	}

	switch algorithm {
	case loadbalancer.AlgorithmRoundRobin:
		return NewRoundRobinBalancer(), nil

	case loadbalancer.AlgorithmLeastConnections:
		return NewLeastConnectionsBalancer(), nil

	case loadbalancer.AlgorithmWeightedRoundRobin:
		return NewWeightedRoundRobinBalancer(options.Weights), nil

	case loadbalancer.AlgorithmLatencyAware:
		return NewLatencyAwareBalancer(), nil

	case loadbalancer.AlgorithmRandom:
		return NewRandomBalancer(), nil

	default:
		return nil, fmt.Errorf("unknown algorithm: %s", algorithm)
	}
}
