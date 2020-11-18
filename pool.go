package service

import (
	"errors"
)

// ErrPoolAlreadyRunning is returned when a function is called and the pool is already running.
var ErrPoolAlreadyRunning = errors.New("cannot add service, pool already running")

// Pool is a handler for multiple services at once. It will run services in parallel in goroutines and terminate all
//      services once a single one has exited.
type Pool interface {
	Service

	// Add inserts a service into the service pool. Returns an error if called after Run.
	Add(s Service) error
}
