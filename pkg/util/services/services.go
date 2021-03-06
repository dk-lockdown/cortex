package services

import (
	"context"
	"time"
)

// Initializes basic service as an "idle" service -- it doesn't do anything in its Running state,
// but still supports all state transitions.
func NewIdleService(up StartingFn, down StoppingFn) Service {
	run := func(ctx context.Context) error {
		<-ctx.Done()
		return nil
	}

	return NewBasicService(up, run, down)
}

// One iteration of the timer service. Called repeatedly until service is stopped, or this function returns error
// in which case, service will fail.
type OneIteration func(ctx context.Context) error

// Runs iteration function on every interval tick. When iteration returns error, service fails.
func NewTimerService(interval time.Duration, start StartingFn, iter OneIteration, stop StoppingFn) Service {
	run := func(ctx context.Context) error {
		t := time.NewTicker(interval)
		defer t.Stop()

		for {
			select {
			case <-t.C:
				err := iter(ctx)
				if err != nil {
					return err
				}

			case <-ctx.Done():
				return nil
			}
		}
	}

	return NewBasicService(start, run, stop)
}

// NewListener provides a simple way to build service listener from supplied functions.
// Functions are only called when not nil.
func NewListener(starting, running func(), stopping, terminated func(from State), failed func(from State, failure error)) Listener {
	return &funcBasedListener{
		startingFn:   starting,
		runningFn:    running,
		stoppingFn:   stopping,
		terminatedFn: terminated,
		failedFn:     failed,
	}
}

type funcBasedListener struct {
	startingFn   func()
	runningFn    func()
	stoppingFn   func(from State)
	terminatedFn func(from State)
	failedFn     func(from State, failure error)
}

func (f *funcBasedListener) Starting() {
	if f.startingFn != nil {
		f.startingFn()
	}
}

func (f funcBasedListener) Running() {
	if f.runningFn != nil {
		f.runningFn()
	}
}

func (f funcBasedListener) Stopping(from State) {
	if f.stoppingFn != nil {
		f.stoppingFn(from)
	}
}

func (f funcBasedListener) Terminated(from State) {
	if f.terminatedFn != nil {
		f.terminatedFn(from)
	}
}

func (f funcBasedListener) Failed(from State, failure error) {
	if f.failedFn != nil {
		f.failedFn(from, failure)
	}
}
