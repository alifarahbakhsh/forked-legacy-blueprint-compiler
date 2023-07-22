package stdlib

import (
	"math/rand"
)

type LoadBalancer[T any] struct {
	Clients []T
}

func NewLoadBalancer[T any](clients []T) * LoadBalancer[T] {
	return &LoadBalancer[T]{Clients: clients}
}

func (this *LoadBalancer[T]) PickClient() T {
	// TODO: Support more policies!
	randIndex := rand.Intn(len(this.Clients))
	return this.Clients[randIndex]
}