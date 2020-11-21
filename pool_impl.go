package service

import (
	"context"
	"fmt"
	"sync"
)

type pool struct {
	mutex            *sync.Mutex
	services         []Service
	lifecycles       map[Service]Lifecycle
	serviceStates    map[Service]State
	lifecycleFactory LifecycleFactory
	running          bool
	startupComplete  chan bool
	stopComplete     chan bool
	lastError        error
	stopping         bool
}

func (p *pool) String() string {
	return "Service Pool"
}

func (p *pool) Add(s Service) Lifecycle {
	p.mutex.Lock()
	if p.running {
		panic("bug: pool already running, cannot add service")
	}
	defer p.mutex.Unlock()
	l := p.lifecycleFactory.Make(s)
	l.OnStateChange(p.onStateChange)
	p.services = append(p.services, s)
	p.lifecycles[s] = l
	return l
}

func (p *pool) Run(lifecycle Lifecycle) error {
	p.mutex.Lock()
	if p.running {
		p.mutex.Unlock()
		panic("bug: pool already running, cannot run again")
	}
	p.running = true
	p.stopping = false
	p.startupComplete = make(chan bool, 1)
	p.stopComplete = make(chan bool, 1)
	p.mutex.Unlock()
	defer func() {
		p.mutex.Lock()
		p.running = false
		p.mutex.Unlock()
	}()
	lifecycle.Starting()

	p.mutex.Lock()
	if p.serviceStates == nil {
		p.serviceStates = map[Service]State{}
	}
	p.mutex.Unlock()

	for _, service := range p.services {
		p.runService(service)
	}

	stopped := false
	startedServices := len(p.services)
waitForStart:
	for i := 0; i < len(p.services); i++ {
		select {
		case <-p.startupComplete:
		case <-p.stopComplete:
			stopped = true
			startedServices--
			break waitForStart
		}
	}

	if !stopped {
		lifecycle.Running()

		select {
		case <-p.stopComplete:
			// One service stopped, initiate shutdown
			startedServices--
		case <-lifecycle.Context().Done():
			p.mutex.Lock()
			p.triggerStop(lifecycle.ShutdownContext())
			p.mutex.Unlock()
		}
	} else {
		p.triggerStop(context.Background())
	}

	for i := 0; i < startedServices; i++ {
		<-p.stopComplete
	}
	if p.lastError != nil {
		lifecycle.Crashed(p.lastError)
		return p.lastError
	}
	lifecycle.Stopped()
	return nil
}

func (p *pool) runService(service Service) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				p.mutex.Lock()
				p.lifecycles[service].Crashed(fmt.Errorf("service crashed (%v)", err))
				p.mutex.Unlock()
			}
		}()
		if err := service.Run(p.lifecycles[service]); err != nil {
			p.lifecycles[service].Crashed(err)
		}
	}()
}

func (p *pool) onStateChange(s Service, l Lifecycle, newState State) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if s == p {
		return
	}

	oldState := p.serviceStates[p]
	p.serviceStates[s] = newState

	if oldState == newState {
		return
	}

	switch newState {
	case StateStarting:
		return
	case StateRunning:
		select {
		case p.startupComplete <- true:
		default:
		}
	case StateStopping:
		p.triggerStop(context.Background())
	case StateStopped:
		p.triggerStop(context.Background())
	case StateCrashed:
		p.lastError = l.Error()
		p.triggerStop(context.Background())
	}
}

func (p *pool) triggerStop(shutdownContext context.Context) {
	if p.stopping {
		return
	}
	p.stopping = true
	for _, s := range p.services {
		l := p.lifecycles[s]
		l.Stop(shutdownContext)
	}
}
