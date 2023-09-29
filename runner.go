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

// coreRunner implements the common functionality for Runner implementations.
// This type is necessary in part so that options don't have to be generics.
type coreRunner struct {
	factory     PolicyFactory
	shouldRetry func(error) bool
	onAttempts  []OnAttempt
	timer       func(time.Duration) (<-chan time.Time, func() bool)
}

// newPolicy creates the retry policy described by these Options.
// If no PolicyFactory is set, this method returns never{}.
func (cr coreRunner) newPolicy(ctx context.Context) Policy {
	if cr.factory == nil {
		taskCtx, cancel := context.WithCancel(ctx)
		return never{
			ctx:    taskCtx,
			cancel: cancel,
		}
	}

	return cr.factory.NewPolicy(ctx)
}

// handleAttempt deals with the aftermath of a task attempt, whether success or fail.
// If onAttempt is set, it is invoked with an Attempt.  If the policy and the error
// allow retries to continue, then interval will be positive and shouldRetry will be true.
func (cr coreRunner) handleAttempt(p Policy, retries int, err error) (interval time.Duration, shouldRetry bool) {
	a := Attempt{
		Context: p.Context(),
		Err:     err,
		Retries: retries,
	}

	shouldRetry = ShouldRetry(err, cr.shouldRetry)

	// slight optimization: if the error indicated no further retries, then there's no
	// reason to consult the policy
	if shouldRetry {
		interval, shouldRetry = p.Next()
		a.Next = interval
	}

	for _, f := range cr.onAttempts {
		f(a)
	}

	return
}

// awaitRetry waits out the interval before returning.  If taskCtx is canceled while waiting,
// this method returns taskCtx.Err().  Otherwise, this method return nil and the next retry
// may continue.
func (cr coreRunner) awaitRetry(taskCtx context.Context, interval time.Duration) (err error) {
	ch, stop := cr.timer(interval)
	select {
	case <-taskCtx.Done():
		err = taskCtx.Err()
		stop()

	case <-ch:
		// time for the next retry
	}

	return
}

// RunnerOption is a configurable option for creating a task runner.
type RunnerOption func(*coreRunner) error

// WithPolicyFactory returns a RunnerOption that assigns the given PolicyFactory
// to the created task runner.
//
// Config in this package implements PolicyFactory.
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

// WithOnAttempt appends one or more callbacks for task results.  This option
// can be applied repeatedly, and the set of OnAttempt callbacks is cumulative.
func WithOnAttempt(fns ...OnAttempt) RunnerOption {
	return func(cr *coreRunner) error {
		cr.onAttempts = append(cr.onAttempts, fns...)
		return nil
	}
}

// Runner is a task executor that honors retry semantics.  A Runner is associated
// with a PolicyFactory, a ShouldRetry strategy, and one or more OnAttempt callbacks.
type Runner[V any] interface {
	// Run executes a task at least once, retrying failures according to
	// the configured PolicyFactory.
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
	coreRunner
}

func (r *runner[V]) Run(parentCtx context.Context, task Task[V]) (result V, err error) {
	p := r.newPolicy(parentCtx)
	defer p.Cancel()
	for taskCtx, retries := p.Context(), 0; taskCtx.Err() == nil; retries++ {
		result, err = task(taskCtx)
		interval, keepTrying := r.handleAttempt(p, retries, err)
		if !keepTrying {
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
func NewRunner[V any](opts ...RunnerOption) (Runner[V], error) {
	r := &runner[V]{
		coreRunner: coreRunner{
			timer: defaultTimer,
		},
	}

	for _, o := range opts {
		if err := o(&r.coreRunner); err != nil {
			return nil, err
		}
	}

	return r, nil
}
