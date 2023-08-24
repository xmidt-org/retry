package retry

import "context"

// Task is the basic type of closure that can be retried in the face of errors.
// This package provides a few convenient ways of coercing functions with
// certain signatures into a Task that returns results of an arbitrary type V.
type Task[V any] func(context.Context) (V, error)

// RunnableTask is the underlying type of closure that implements a task.
// A RunnableTask may or may not accept a context, but will always return
// a result and an error.
type RunnableTask[V any] interface {
	~func() (V, error) | ~func(context.Context) (V, error)
}

// AsTask normalizes any RunnableTask into a Task so that it can be submitted
// to a Runner.
func AsTask[V any, RT RunnableTask[V]](rt RT) Task[V] {
	if t, ok := any(rt).(func(context.Context) (V, error)); ok {
		return t
	}

	t := any(rt).(func() (V, error))
	return func(_ context.Context) (V, error) {
		return t()
	}
}

// SimpleTask is the underlying type for any task closure which doesn't return
// a custom value.  Tasks of this type must be wrapped with a value, either in calling
// code or via one of the convenience functions in this package.
type SimpleTask interface {
	~func() error | ~func(context.Context) error
}

// normalizeSimpleTask is performs the common work of coercing a SimpleTask
// into a known closure that always accepts a context.
func normalizeSimpleTask[T SimpleTask](raw T) func(context.Context) error {
	if task, ok := any(raw).(func(context.Context) error); ok {
		return task
	}

	f := any(raw).(func() error)
	return func(context.Context) error {
		return f()
	}
}

// AddZero wraps a SimpleTask so that it always returns the zero value of some arbitrary type V,
// regardless of any error.
//
// The returned task closure may be passed to a Runner[V].
func AddZero[V any, T SimpleTask](raw T) func(context.Context) (V, error) {
	task := normalizeSimpleTask(raw)
	return func(ctx context.Context) (V, error) {
		var result V // zero value of type V
		return result, task(ctx)
	}
}

// AddValue wraps a SimpleTask so that it always returns the given result, regardless of
// any error.
//
// The returned task closure may be passed to a Runner[V].
func AddValue[V any, T SimpleTask](result V, raw T) func(context.Context) (V, error) {
	task := normalizeSimpleTask(raw)
	return func(ctx context.Context) (V, error) {
		return result, task(ctx)
	}
}

// AddSuccessFail wraps a SimpleTask so that it returns a value based on whether an error
// is encountered.  This function is useful when invoking functions from other packages
// that only return errors.
//
// The returned task closure may be passed to a Runner[V].
func AddSuccessFail[V any, T SimpleTask](success, fail V, raw T) func(context.Context) (V, error) {
	task := normalizeSimpleTask(raw)
	return func(ctx context.Context) (result V, err error) {
		if err = task(ctx); err == nil {
			result = success
		} else {
			result = fail
		}

		return
	}
}
