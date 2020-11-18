package service

import (
	"context"
	"sync"

	"github.com/containerssh/log"
)

type pool struct {
	onReady      []func()
	onShutdown   []func()
	startWg      *sync.WaitGroup
	finishWg     *sync.WaitGroup
	mutex        *sync.Mutex
	running      bool
	services     []Service
	err          error
	shuttingDown bool
	logger       log.Logger
	readyHandled map[Service]bool
}

func (p *pool) getServiceReadyHandler(service Service, errorState bool) func() {
	return func() {
		p.mutex.Lock()
		defer p.mutex.Unlock()
		if val, ok := p.readyHandled[service]; val && ok {
			return
		}
		p.startWg.Done()
		p.readyHandled[service] = true
		if !errorState {
			p.logger.Debugf("%s is now ready", service.String())
		}
	}
}

func (p *pool) OnReady(f func()) {
	if f == nil {
		panic("tried to call OnReady with a nil handler")
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()
	if p.running {
		panic("tried to call OnReady after the pool was already started")
	}
	p.onReady = append(p.onReady, f)
}

func (p *pool) OnShutdown(f func()) {
	if f == nil {
		panic("tried to call OnShutdown with a nil handler")
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()
	if p.running {
		panic("tried to call OnShutdown after the pool was already started")
	}
	p.onShutdown = append(p.onShutdown, f)
}

func (p *pool) Wait() error {
	p.finishWg.Wait()
	return p.err
}

func (p *pool) runService(service Service) {
	defer func() {
		p.finishWg.Done()
		if r := recover(); r != nil {
			p.logger.Criticalf("%s panicked, shutting down (%v)", service.String(), r)
			p.mutex.Lock()
			defer p.mutex.Unlock()
			p.triggerShutdown(context.Background())
		}
	}()

	err := service.Run()
	p.getServiceReadyHandler(service, true)()

	p.mutex.Lock()
	defer p.mutex.Unlock()
	if err != nil {
		p.logger.Errorf("%s exited with an error (%v)", service.String(), err)
		p.err = err
	}
	p.triggerShutdown(context.Background())
}

func (p *pool) Run() error {
	p.mutex.Lock()
	if p.running {
		p.mutex.Unlock()
		return p.Wait()
	}
	p.startWg.Add(len(p.services))
	p.finishWg.Add(len(p.services))
	for _, s := range p.services {
		go p.runService(s)
	}
	p.mutex.Unlock()

	p.startWg.Wait()
	if p.err == nil {
		for _, onReady := range p.onReady {
			onReady()
		}
	}

	err := p.Wait()
	p.mutex.Lock()
	p.running = false
	p.shuttingDown = false
	p.mutex.Unlock()
	return err
}

// triggerShutdown triggers a pool shutdown if it is running and shuttingDown has been triggered. This method must
//                 only be called within the mutex lock.
func (p *pool) triggerShutdown(shutdownContext context.Context) {
	if !p.running || p.shuttingDown {
		return
	}
	p.logger.Debugf("pool shutting down")
	p.shuttingDown = true
	for _, onShutdownFunc := range p.onShutdown {
		onShutdownFunc()
	}
	for _, s := range p.services {
		go s.Shutdown(shutdownContext)
	}
}

func (p *pool) Shutdown(shutdownContext context.Context) {
	p.mutex.Lock()
	p.triggerShutdown(shutdownContext)
	p.mutex.Unlock()
	_ = p.Wait()
}

func (p *pool) Add(s Service) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.running {
		return ErrPoolAlreadyRunning
	}
	s.OnReady(p.getServiceReadyHandler(s, false))
	p.services = append(p.services, s)
	return nil
}

func (p *pool) String() string {
	return "service pool"
}
