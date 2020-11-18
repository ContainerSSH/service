package service

import (
	"context"
)

// Service is an interface that specifies the minimum requirements for the server.
type Service interface {
	// OnReady adds a function handler to be called when the service is ready to serve user requests. Calling this
	//         method again adds a second function to be called. Must be called before Run.
	OnReady(func())

	// OnShutdown adds a function handler to be called just before the service is starting to shut down. This can be
	//            used to remove the service from a load balancer. Calling this method again adds a second function to
	//            be called. Must be called before Run.
	OnShutdown(func())

	// Wait waits for the service pool to complete, then returns the error if any. If the pool is not running it returns
	//      immediately.
	Wait() error

	// Run is called to execute this service. It returns when the service has finished. It only returns an error if the
	//     service finished abnormally. Calling this method a second time will lead to the same behavior as Wait
	Run() error

	// Shutdown requests that the service shut down within the context provided in shutdownContext. It returns when the
	//          service has finished. If the service is not running it returns immediately.
	Shutdown(shutdownContext context.Context)

	// String should return a user-readable name for the service.
	String() string
}
