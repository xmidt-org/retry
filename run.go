package retry

import (
	"context"
	"time"
)

// Options describes how (or if) a task should be retried.
type Options struct {
	// Factory is an optional PolicyFactory that is used to create a Policy
	// each time a task is run.  If this field is unset, no retries are done.
	Factory PolicyFactory

	// ShouldRetry is an optional predicate for determining whether an error should
	// prevent further retries.  If this field is unset, then either the error is consulted
	// (via the ShouldRetryable interface) or a retry is prevented.
	ShouldRetry func(error) bool

	// OnFail is an optional closure that is invoked each time a task fails, including
	// the first time.  This function is passed the interval for the current retry that
	// was just done.  If this is the first time the task was attempted, this closure
	// is passed a zero (0) duration.
	OnFail func(error, time.Duration)

	sleep func(time.Duration)
}

// newPolicy creates the retry policy described by these Options.
// If no PolicyFactory is set, this method returns never{}.
func (o Options) newPolicy() Policy {
	if o.Factory == nil {
		return never{}
	}

	return o.Factory.NewPolicy()
}

// handleTaskError examines the error returned by a task to determine
// whether retries should continue.  This method is also passed the duration
// of the previous retry (zero for the first time), and will dispatch
// to any configured OnFail closure as appropriate.
func (o Options) handleTaskError(err error, d time.Duration) (shouldRetry bool) {
	if err != nil && o.OnFail != nil {
		o.OnFail(err, d)
	}

	shouldRetry = ShouldRetry(err, o.ShouldRetry)
	return
}

// doSleep handles advancing to the next interval and sleeping as appropriate.
func (o Options) doSleep(p Policy) (next time.Duration, ok bool) {
	next, ok = p.Next()
	if ok {
		f := o.sleep
		if f == nil {
			f = time.Sleep
		}

		f(next)
	}

	return
}

// Run executes a task at least once, optionally retrying it on failure
// as described in the given Options.
func Run[T ~func() error](o Options, task T) (err error) {
	var (
		p            = o.newPolicy()
		interval     time.Duration
		keepRetrying = true
	)

	for keepRetrying {
		err = task()
		if !o.handleTaskError(err, interval) {
			break
		}

		interval, keepRetrying = o.doSleep(p)
	}

	return
}

// RunWithData executes a task at least once, optionally retrying it on failure
// as described in the given Options.
//
// The task passed to this function returns an arbitrary value V, and that value
// is returned by this function upon success.
func RunWithData[V any, T ~func() (V, error)](o Options, task T) (result V, err error) {
	var (
		p            = o.newPolicy()
		interval     time.Duration
		keepRetrying = true
	)

	for keepRetrying {
		result, err = task()
		if !o.handleTaskError(err, interval) {
			break
		}

		interval, keepRetrying = o.doSleep(p)
	}

	return
}

// RunCtx executes a task at least once, optionally retrying it on failure
// as described in the given Options.  The given context's cancelation
// semantics are honored by this function itself, in addition to passing
// that context with each invocation of the task.
func RunCtx[T ~func(context.Context) error](ctx context.Context, o Options, task T) (err error) {
	var (
		p            = o.newPolicy()
		interval     time.Duration
		keepRetrying = true
	)

	for err = ctx.Err(); keepRetrying && err == nil; err = ctx.Err() {
		err = task(ctx)
		if !o.handleTaskError(err, interval) {
			break
		} else if err = ctx.Err(); err != nil {
			break
		}

		interval, keepRetrying = o.doSleep(p)
	}

	return
}

// RunWithDataCtx executes a task at least once, optionally retrying it on failure
// as described in the given Options.  The given context's cancelation
// semantics are honored by this function itself, in addition to passing
// that context with each invocation of the task.
//
// The task passed to this function returns an arbitrary value V, and that value
// is returned by this function upon success.
func RunWithDataCtx[V any, T ~func(context.Context) (V, error)](ctx context.Context, o Options, task T) (result V, err error) {
	var (
		p            = o.newPolicy()
		interval     time.Duration
		keepRetrying = true
	)

	for err = ctx.Err(); keepRetrying && err == nil; err = ctx.Err() {
		result, err = task(ctx)
		if !o.handleTaskError(err, interval) {
			break
		} else if err = ctx.Err(); err != nil {
			break
		}

		interval, keepRetrying = o.doSleep(p)
	}

	return
}
