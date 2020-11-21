package service_test

import (
	"context"
	"testing"

	"github.com/containerssh/log/standard"
	"github.com/stretchr/testify/assert"

	"github.com/containerssh/service"
)

type testService struct {
}

func (t *testService) String() string {
	return "Test service"
}

func (t *testService) Run(lifecycle service.Lifecycle) error {
	lifecycle.Starting()
	lifecycle.Running()
	<-lifecycle.Context().Done()
	lifecycle.Stopping()
	lifecycle.Stopped()
	return nil
}

func TestLifecycle(t *testing.T) {
	logger := standard.New()
	l := service.NewLifecycle(&testService{}, logger)
	starting := make(chan bool)
	running := make(chan bool)
	stopping := make(chan bool)
	stopped := make(chan bool)
	stopExited := make(chan bool)
	l.OnStarting(func(s service.Service, l service.Lifecycle) {
		starting <- true
	})
	l.OnRunning(func(s service.Service, l service.Lifecycle) {
		running <- true
	})
	l.OnStopping(func(s service.Service, l service.Lifecycle, shutdownContext context.Context) {
		stopping <- true
	})
	l.OnStopped(func(s service.Service, l service.Lifecycle) {
		stopped <- true
	})
	go func() {
		if err := l.Run(); err != nil {
			assert.Fail(t, "service crashed", err)
		}
	}()
	<-starting
	<-running
	go func() {
		l.Stop(context.Background())
		stopExited <- true
	}()
	<-stopping
	<-stopped
	<-stopExited
}
