package service

import (
	"sync"

	"github.com/containerssh/log"
)

// NewPool creates a new service pool that can be used to run and manage multiple services in parallel.
func NewPool(lifecycleFactory LifecycleFactory, logger log.Logger) Pool {
	return &pool{
		mutex:            &sync.Mutex{},
		services:         []Service{},
		lifecycleFactory: lifecycleFactory,
		lifecycles:       map[Service]Lifecycle{},
	}
}
