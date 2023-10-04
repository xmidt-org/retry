// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package retry

import (
	"context"
	"time"
)

// defaultTimer is the strategy used to create a timer using the stdlib.
func defaultTimer(d time.Duration) (<-chan time.Time, func() bool) {
	t := time.NewTimer(d)
	return t.C, t.Stop
}

// RunnerOption is a configurable option for creating a task runner.
type RunnerOption[V any] interface {
	apply(*runner[V]) error
}

type runnerOptionFunc[V any] func(*runner[V]) error

func (rof runnerOptionFunc[V]) apply(r *runner[V]) error { return rof(r) }

// WithPolicyFactory returns a RunnerOption that assigns the given PolicyFactory
// to the created task runner.
//
// Config in this package implements PolicyFactory.
func WithPolicyFactory[V any](pf PolicyFactory) RunnerOption[V] {
	return runnerOptionFunc[V](func(r *runner[V]) error {
		r.factory = pf
		return nil
	})
}

// WithShouldRetry adds a predicate to the created task runner that will be
// used to determine if an error should be retried or should halt further attempts.
// This predicate is used if the error itself does not expose retryablity semantics
// via a ShouldRetry method.
func WithShouldRetry[V any](sr ShouldRetry[V]) RunnerOption[V] {
	return runnerOptionFunc[V](func(r *runner[V]) error {
		r.shouldRetry = sr
		return nil
	})
}

// WithOnAttempt appends one or more callbacks for task results.  This option
// can be applied repeatedly, and the set of OnAttempt callbacks is cumulative.
func WithOnAttempt[V any](fns ...OnAttempt[V]) RunnerOption[V] {
	return runnerOptionFunc[V](func(r *runner[V]) error {
		r.onAttempts = append(r.onAttempts, fns...)
		return nil
	})
}

// Runner is a task executor that honors retry semantics.  A Runner is associated
// with a PolicyFactory, a ShouldRetry strategy, and one or more OnAttempt callbacks.
type Runner[V any] interface {
	// Run executes a task at least once, retrying failures according to
	// the configured PolicyFactory.  If all attempts fail, this method returns
	// the error along with the zero value for V.  Otherwise, the value V that
	// resulted from the final, successful attempt is returned along with a nil error.
	//
	// The context passed to this method must never be nil.  Use context.Background()
	// or context.TODO() as appropriate rather than nil.
	//
	// The configured PolicyFactory may impose a time limit, e.g. the Config.MaxElapsedTime
	// field.  In this case, if the time limit is reached, task attempts will halt regardless
	// of the state of the parent context.
	Run(context.Context, Task[V]) (V, error)
}

type runner[V any] struct {
	factory     PolicyFactory
	shouldRetry ShouldRetry[V]
	onAttempts  []OnAttempt[V]
	timer       func(time.Duration) (<-chan time.Time, func() bool)
}

// newPolicy creates a Policy for a series of attempts.
func (r *runner[V]) newPolicy(ctx context.Context) Policy {
	if r.factory == nil {
		taskCtx, cancel := context.WithCancel(ctx)
		return &never{
			ctx:    taskCtx,
			cancel: cancel,
		}
	}

	return r.factory.NewPolicy(ctx)
}

// handleAttempt deals with the aftermath of a task attempt, whether success or fail.
// If onAttempt is set, it is invoked with an Attempt.  If the policy and the error
// allow retries to continue, then interval will be positive and shouldRetry will be true.
func (r *runner[V]) handleAttempt(p Policy, retries int, result V, err error) (interval time.Duration, shouldRetry bool) {
	a := Attempt[V]{
		Context: p.Context(),
		Result:  result,
		Err:     err,
		Retries: retries,
	}

	shouldRetry = CheckRetry(result, err, r.shouldRetry)

	// slight optimization: if the error indicated no further retries, then there's no
	// reason to consult the policy
	if shouldRetry {
		interval, shouldRetry = p.Next()
		a.Next = interval
	}

	for _, f := range r.onAttempts {
		f(a)
	}

	return
}

// awaitRetry waits out the interval before returning.  If taskCtx is canceled while waiting,
// this method returns taskCtx.Err().  Otherwise, this method return nil and the next retry
// may continue.
func (r *runner[V]) awaitRetry(taskCtx context.Context, interval time.Duration) (err error) {
	ch, stop := r.timer(interval)
	select {
	case <-taskCtx.Done():
		err = taskCtx.Err()
		stop()

	case <-ch:
		// time for the next retry
	}

	return
}

func (r *runner[V]) Run(parentCtx context.Context, task Task[V]) (result V, err error) {
	p := r.newPolicy(parentCtx)
	defer p.Cancel()

	var attemptResult V
	for taskCtx, retries := p.Context(), 0; taskCtx.Err() == nil; retries++ {
		attemptResult, err = task(taskCtx)
		interval, keepTrying := r.handleAttempt(p, retries, attemptResult, err)
		if !keepTrying {
			result = attemptResult
			break
		}

		err = r.awaitRetry(taskCtx, interval)
		if err != nil {
			break
		}
	}

	return
}

// NewRunner creates a Runner using the supplied set of options.
func NewRunner[V any](opts ...RunnerOption[V]) (Runner[V], error) {
	r := &runner[V]{
		timer: defaultTimer,
	}

	for _, o := range opts {
		if err := o.apply(r); err != nil {
			return nil, err
		}
	}

	return r, nil
}
