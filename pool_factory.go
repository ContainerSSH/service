package service

import (
	"github.com/containerssh/log"
)

// NewPool creates a new service pool that can be used to run and manage multiple services in parallel.
func NewPool(logger log.Logger) Pool {
	return &pool{}
}
