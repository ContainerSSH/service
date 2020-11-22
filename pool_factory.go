package service

import (
	"sync"
)

// NewPool creates a new service pool that can be used to run and manage multiple services in parallel.
func NewPool(lifecycleFactory LifecycleFactory) Pool {
	return &pool{
		mutex:            &sync.Mutex{},
		services:         []Service{},
		lifecycleFactory: lifecycleFactory,
		lifecycles:       map[Service]Lifecycle{},
		serviceStates:    map[Service]State{},
		startupComplete:  make(chan struct{}, 1),
		stopComplete:     make(chan struct{}, 1),
	}
}
