package service

// Service is an interface that specifies the minimum requirements for a generic service.
type Service interface {
	// String should return a user-readable name for the service.
	String() string

	// Run is called to execute this service. It returns when the service has finished. It only returns an error if the
	// service finished abnormally. The Run implementation must observe the context in the lifecycle and call the
	// appropriate hooks as it enters the stages of its life.
	//
	// During implementation the Run method must implement the following steps:
	//
	// - When the service starts it must immediately call lifecycle.Starting().
	// - When the service is ready to serve user requests it must call lifecycle.Running().
	// - During the running phase the service should regularly check lifecycle.ShouldStop() or use lifecycle.Context()
	//   to determine if it should stop execution.
	// - When the service begins to shut down it must call lifecycle.Stopping(), which will return a shutdown context.
	//   The shutdown context gives the service a deadline by which to gracefully shut down.
	// - When the shutdown context expires the service must abort graceful shutdown and stop as soon as possible.
	// - When the shutdown is complete the service must call `lifecycle.Stopped()`.
	// - If, at any point the service exits with an error it must call `lifecycle.Crashed(err)`.
	// - The service should implement `recover()` to handle recovering from a `panic()` and call lifecycle.Crashed(err).
	Run(lifecycle Lifecycle) error
}
