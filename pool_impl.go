package service

import (
	"context"
	"sync"
)

type pool struct {
	mutex            *sync.Mutex
	services         []Service
	lifecycles       map[Service]Lifecycle
	serviceStates    map[Service]State
	lifecycleFactory LifecycleFactory
	running          bool
	startupComplete  chan struct{}
	stopComplete     chan struct{}
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

func (p *pool) RunWithLifecycle(lifecycle Lifecycle) error {
	p.mutex.Lock()
	if p.running {
		p.mutex.Unlock()
		panic("bug: pool already running, cannot run again")
	}
	p.running = true
	p.stopping = false
	p.startupComplete = make(chan struct{}, 1)
	p.stopComplete = make(chan struct{}, 1)
	p.mutex.Unlock()
	defer func() {
		p.mutex.Lock()
		p.running = false
		p.mutex.Unlock()
	}()

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
	finished := false
	for i := 0; i < len(p.services); i++ {
		select {
		case <-p.startupComplete:
		case <-p.stopComplete:
			stopped = true
			startedServices--
			finished = true
		}
		if finished {
			break
		}
	}

	if !stopped {
		lifecycle.Running()

		select {
		case <-p.stopComplete:
			// One service stopped, initiate shutdown
			startedServices--
		case <-lifecycle.Context().Done():
			lifecycle.Stopping()
			p.triggerStop(lifecycle.ShutdownContext())
		}
	} else {
		p.triggerStop(context.Background())
	}

	for i := 0; i < startedServices; i++ {
		<-p.stopComplete
	}
	return p.lastError
}

func (p *pool) runService(service Service) {
	go func() {
		_ = p.lifecycles[service].Run()
	}()
}

func (p *pool) onStateChange(s Service, l Lifecycle, newState State) {
	if s == p {
		return
	}

	p.mutex.Lock()
	oldState := p.serviceStates[p]
	p.serviceStates[s] = newState
	p.mutex.Unlock()

	if oldState == newState {
		return
	}

	switch newState {
	case StateStarting:
		return
	case StateRunning:
		select {
		case p.startupComplete <- struct{}{}:
		default:
		}
	case StateStopping:
		p.triggerStop(context.Background())
	case StateStopped:
		p.triggerStop(context.Background())
		p.stopComplete <- struct{}{}
	case StateCrashed:
		p.lastError = l.Error()
		p.triggerStop(context.Background())
		p.stopComplete <- struct{}{}
	}
}

func (p *pool) triggerStop(shutdownContext context.Context) {
	p.mutex.Lock()
	if p.stopping {
		p.mutex.Unlock()
		return
	}
	p.stopping = true
	svc := p.services
	p.mutex.Unlock()

	wg := &sync.WaitGroup{}
	wg.Add(len(svc))
	for _, s := range svc {
		l := p.lifecycles[s]
		go func() {
			defer wg.Done()
			l.Stop(shutdownContext)
		}()
	}
	wg.Wait()
}
