package semaphore

import (
	"errors"
	"fmt"
	"os"

	"github.com/tedsuo/ifrit"
)

type Resource interface {
	Release()
}

type Semaphore interface {
	ifrit.Runner

	Acquire() (Resource, error)
}

type request chan struct{}

type resource struct {
	inflightRequests chan request
}

type semaphore struct {
	inflightRequests chan request
	pendingRequests  chan request
	maxPending       int
}

func New(maxInflight, maxPending int) Semaphore {
	return &semaphore{
		inflightRequests: make(chan request, maxInflight),
		pendingRequests:  make(chan request, maxPending),
		maxPending:       maxPending,
	}
}

func (os *semaphore) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	var pendingRequest request

	close(ready)

	for {
		select {
		case pendingRequest = <-os.pendingRequests:
		case <-signals:
			return nil
		}

		select {
		case os.inflightRequests <- pendingRequest:
			close(pendingRequest)
		case <-signals:
			return nil
		}
	}

	return nil
}
func (os *semaphore) Acquire() (Resource, error) {
	var pendingRequest request = make(chan struct{})
	select {
	case os.pendingRequests <- pendingRequest:
	default:
		return nil, errors.New(fmt.Sprintf("Cannot queue request, maxPending reached: %d", os.maxPending))
	}

	<-pendingRequest

	return &resource{os.inflightRequests}, nil
}

func (r *resource) Release() {
	<-r.inflightRequests
}
