package service

import (
	"sync"
)

// NewPool creates a new service pool that can be used to run and manage multiple services in parallel.
func NewPool() Pool {
	return &pool{
		onReady:    nil,
		onShutdown: nil,
		finishWg:   &sync.WaitGroup{},
		mutex:      &sync.Mutex{},
	}
}
