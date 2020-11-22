package service

import (
	"context"
	"sync"
)

func NewLifecycle(service Service) Lifecycle {
	ctx, cancelFunc := context.WithCancel(context.Background())
	return &lifecycle{
		service:         service,
		state:           StateStopped,
		mutex:           &sync.Mutex{},
		runningContext:  ctx,
		cancelRun:       cancelFunc,
		shutdownContext: context.Background(),
	}
}

func NewLifecycleFactory() LifecycleFactory {
	return &lifecycleFactory{}
}

type LifecycleFactory interface {
	Make(service Service) Lifecycle
}

type lifecycleFactory struct {
}

func (l *lifecycleFactory) Make(service Service) Lifecycle {
	return NewLifecycle(service)
}
