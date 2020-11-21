package service

import (
	"context"
	"sync"

	"github.com/containerssh/log"
)

func NewLifecycle(service Service, logger log.Logger) Lifecycle {
	ctx, cancelFunc := context.WithCancel(context.Background())
	return &lifecycle{
		service:         service,
		logger:          logger,
		state:           StateStopped,
		mutex:           &sync.Mutex{},
		runningContext:  ctx,
		cancelRun:       cancelFunc,
		shutdownContext: context.Background(),
	}
}

func NewLifecycleFactory(logger log.Logger) LifecycleFactory {
	return &lifecycleFactory{
		logger: logger,
	}
}

type LifecycleFactory interface {
	Make(service Service) Lifecycle
}

type lifecycleFactory struct {
	logger log.Logger
}

func (l *lifecycleFactory) Make(service Service) Lifecycle {
	return NewLifecycle(service, l.logger)
}
