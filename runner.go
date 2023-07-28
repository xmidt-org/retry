package retry

import (
	"context"
	"time"
)

// coreRunner implements the common functionality for Runner implementations.
type coreRunner struct {
	factory     PolicyFactory
	shouldRetry func(error) bool
	onFail      func(error, time.Duration)

	sleep func(time.Duration)
}

// newPolicy creates the retry policy described by these Options.
// If no PolicyFactory is set, this method returns never{}.
func (cr coreRunner) newPolicy() Policy {
	if cr.factory == nil {
		return never{}
	}

	return cr.factory.NewPolicy()
}

// handleTaskError examines the error returned by a task to determine
// whether retries should continue.  This method is also passed the duration
// of the previous retry (zero for the first time), and will dispatch
// to any configured OnFail closure as appropriate.
func (cr coreRunner) handleTaskError(err error, d time.Duration) (shouldRetry bool) {
	if err != nil && cr.onFail != nil {
		cr.onFail(err, d)
	}

	shouldRetry = ShouldRetry(err, cr.shouldRetry)
	return
}

// doSleep handles advancing to the next interval and sleeping as appropriate.
func (cr coreRunner) doSleep(p Policy) (next time.Duration, ok bool) {
	next, ok = p.Next()
	if ok {
		cr.sleep(next)
	}

	return
}

// RunnerOption is a configurable option for creating a task runner.
type RunnerOption func(*coreRunner) error

// WithPolicyFactory returns a RunnerOption that assigns the given PolicyFactory
// to the created task runner.
//
// Note: Config in this package implements PolicyFactory.
func WithPolicyFactory(pf PolicyFactory) RunnerOption {
	return func(cr *coreRunner) error {
		cr.factory = pf
		return nil
	}
}

// WithShouldRetry adds a predicate to the created task runner that will be
// used to determine if an error should be retried or should halt further attempts.
// This predicate is used if the error itself does not expose retryablity semantics
// via a ShouldRetry method.
func WithShouldRetry(sr func(error) bool) RunnerOption {
	return func(cr *coreRunner) error {
		cr.shouldRetry = sr
		return nil
	}
}

// WithOnFail adds a callback to the created task runner that will be invoked
// every time a task attempt fails, including the first attempt (before any retries).
//
// On the first attempt, the duration will be zero (0).  For each retry, the duration
// will be the interval of the retry just attempted.
func WithOnFail(of func(error, time.Duration)) RunnerOption {
	return func(cr *coreRunner) error {
		cr.onFail = of
		return nil
	}
}

// Runner is a task executor that honors retry semantics.
type Runner interface {
	// Run executes a task at least once, retrying failures according to
	// the configured PolicyFactory.
	Run(func() error) error

	// RunCtx executes a task at least once, retrying failures according to
	// the configured PolicyFactory.
	//
	// This variant honors context cancelation semantics.
	RunCtx(context.Context, func(context.Context) error) error
}

type runner struct {
	coreRunner
}

func (r runner) Run(task func() error) (err error) {
	var (
		p            = r.newPolicy()
		interval     time.Duration
		keepRetrying = true
	)

	for keepRetrying {
		err = task()
		if !r.handleTaskError(err, interval) {
			break
		}

		interval, keepRetrying = r.doSleep(p)
	}

	return
}

func (r runner) RunCtx(ctx context.Context, task func(context.Context) error) (err error) {
	var (
		p            = r.newPolicy()
		interval     time.Duration
		keepRetrying = true
	)

	for err = ctx.Err(); keepRetrying && err == nil; err = ctx.Err() {
		err = task(ctx)
		if !r.handleTaskError(err, interval) {
			break
		} else if err = ctx.Err(); err != nil {
			break
		}

		interval, keepRetrying = r.doSleep(p)
	}

	return
}

// NewRunner creates a Runner using the supplied set of options.
func NewRunner(opts ...RunnerOption) (Runner, error) {
	r := runner{
		coreRunner: coreRunner{
			sleep: time.Sleep,
		},
	}

	for _, o := range opts {
		if err := o(&r.coreRunner); err != nil {
			return nil, err
		}
	}

	return r, nil
}

// RunnerWithData is a Runner variant that allows tasks to return an
// arbitrary data type.
type RunnerWithData[V any] interface {
	// Run executes a task at least once, retrying failures according to
	// the configured PolicyFactory.
	//
	// The returned type V is the either (1) the successful return of the
	// task, or (2) the zero value of V if all attempts failed.
	Run(func() (V, error)) (V, error)

	// RunCtx executes a task at least once, retrying failures according to
	// the configured PolicyFactory.
	//
	// The returned type V is the either (1) the successful return of the
	// task, or (2) the zero value of V if all attempts failed.
	//
	// This variant honors context cancelation semantics.
	RunCtx(context.Context, func(context.Context) (V, error)) (V, error)
}

type runnerWithData[V any] struct {
	coreRunner
}

func (r runnerWithData[V]) Run(task func() (V, error)) (result V, err error) {
	var (
		p            = r.newPolicy()
		interval     time.Duration
		keepRetrying = true
	)

	for keepRetrying {
		result, err = task()
		if !r.handleTaskError(err, interval) {
			break
		}

		interval, keepRetrying = r.doSleep(p)
	}

	return
}

func (r runnerWithData[V]) RunCtx(ctx context.Context, task func(context.Context) (V, error)) (result V, err error) {
	var (
		p            = r.newPolicy()
		interval     time.Duration
		keepRetrying = true
	)

	for err = ctx.Err(); keepRetrying && err == nil; err = ctx.Err() {
		result, err = task(ctx)
		if !r.handleTaskError(err, interval) {
			break
		} else if err = ctx.Err(); err != nil {
			break
		}

		interval, keepRetrying = r.doSleep(p)
	}

	return
}

// NewRunnerWithData creates a RunnerWithData using the supplied set of options.  All tasks
// executed by the returned runner must return a value of type V in addition to an error.
func NewRunnerWithData[V any](opts ...RunnerOption) (RunnerWithData[V], error) {
	r := runnerWithData[V]{
		coreRunner: coreRunner{
			sleep: time.Sleep,
		},
	}

	for _, o := range opts {
		if err := o(&r.coreRunner); err != nil {
			return nil, err
		}
	}

	return r, nil
}
