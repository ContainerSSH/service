package service_test

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/containerssh/log"
	"github.com/containerssh/log/formatter/ljson"
	"github.com/containerssh/log/pipeline"
	"github.com/stretchr/testify/assert"

	"github.com/containerssh/service"
)

func TestEmptyPool(t *testing.T) {
	logger := pipeline.NewLoggerPipeline(log.LevelInfo, ljson.NewLJsonLogFormatter(), ioutil.Discard)
	pool := service.NewPool(service.NewLifecycleFactory(logger), logger)
	lifecycle := service.NewLifecycle(pool, logger)
	started := make(chan bool)
	stopped := make(chan bool)
	lifecycle.OnRunning(func(s service.Service, l service.Lifecycle) {
		started <- true
	})
	go func() {
		err := lifecycle.Run()
		if err != nil {
			t.Fail()
		}
		stopped <- true
	}()
	<-started
	lifecycle.Stop(context.Background())
	<-stopped
}

func TestOneService(t *testing.T) {
	logger := pipeline.NewLoggerPipeline(log.LevelInfo, ljson.NewLJsonLogFormatter(), ioutil.Discard)
	pool := service.NewPool(service.NewLifecycleFactory(logger), logger)
	poolLifecycle := service.NewLifecycle(pool, logger)
	poolStarted := make(chan bool)
	poolStopped := make(chan bool)
	var poolStates []service.State
	var serviceStates []service.State
	poolLifecycle.OnRunning(func(s service.Service, l service.Lifecycle) {
		poolStarted <- true
	})
	poolLifecycle.OnStateChange(func(s service.Service, l service.Lifecycle, state service.State) {
		poolStates = append(poolStates, state)
	})

	s := newTestService()
	pool.Add(s).OnStateChange(func(s service.Service, l service.Lifecycle, state service.State) {
		serviceStates = append(serviceStates, state)
	})

	go func() {
		err := poolLifecycle.Run()
		if err != nil {
			t.Fail()
		}
		poolStopped <- true
	}()

	<-poolStarted
	poolLifecycle.Stop(context.Background())
	assert.Equal(t, []service.State{
		service.StateStarting,
		service.StateRunning,
		service.StateStopping,
		service.StateStopped,
	}, serviceStates)
	assert.Equal(t, []service.State{
		service.StateStarting,
		service.StateRunning,
		service.StateStopping,
		service.StateStopped,
	}, poolStates)
}
