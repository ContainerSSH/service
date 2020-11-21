package service_test

import (
	"errors"

	"github.com/containerssh/service"
)

type testService struct {
	crash chan bool
}

func (t *testService) String() string {
	return "Test service"
}

func (t *testService) RunWithLifecycle(lifecycle service.Lifecycle) error {
	lifecycle.Running()
	ctx := lifecycle.Context()
	select {
	case <-ctx.Done():
		lifecycle.Stopping()
		return nil
	case <-t.crash:
		return errors.New("crash")
	}
}

func (t *testService) Crash() {
	select {
	case t.crash <- true:
	default:
	}
}

func newTestService() *testService {
	return &testService{
		crash: make(chan bool, 1),
	}
}
