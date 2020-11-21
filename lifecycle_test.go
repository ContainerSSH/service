package service_test

import (
	"context"
	"testing"

	"github.com/containerssh/log/standard"
	"github.com/stretchr/testify/assert"

	"github.com/containerssh/service"
)

func TestLifecycle(t *testing.T) {
	logger := standard.New()
	l := service.NewLifecycle(newTestService(), logger)
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
	l.OnCrashed(func(s service.Service, l service.Lifecycle, err error) {
		t.Fail()
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

func TestCrash(t *testing.T) {
	logger := standard.New()
	s := newTestService()
	l := service.NewLifecycle(s, logger)
	starting := make(chan bool)
	running := make(chan bool)
	crashed := make(chan bool, 1)
	l.OnStarting(func(s service.Service, l service.Lifecycle) {
		starting <- true
	})
	l.OnRunning(func(s service.Service, l service.Lifecycle) {
		running <- true
	})
	l.OnCrashed(func(s service.Service, l service.Lifecycle, err error) {
		crashed <- true
	})
	l.OnStopped(func(s service.Service, l service.Lifecycle) {
		t.Fail()
	})
	l.OnStopping(func(s service.Service, l service.Lifecycle, shutdownContext context.Context) {
		t.Fail()
	})
	go func() {
		if err := l.Run(); err == nil {
			assert.Fail(t, "service did not crash")
		}
	}()
	<-starting
	<-running
	s.Crash()
	<-crashed
}
