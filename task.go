package retry

import "context"

// Task is the basic type of closure that can be retried in the face of errors.
type Task[V any] func(context.Context) (V, error)

// RunnableTask is the underlying type of closure that implements a task.
type RunnableTask[V any] interface {
	~func() (V, error) | ~func(context.Context) (V, error)
}

// AsTask coerces any RunnableTask into a Task so that it can be submitted
// to a Runner.
func AsTask[V any, RT RunnableTask[V]](rt RT) Task[V] {
	if t, ok := any(rt).(Task[V]); ok {
		return t
	}

	t := any(rt).(func() (V, error))
	return func(_ context.Context) (V, error) {
		return t()
	}
}

// SimpleTask is the underlying type for any task closure which doesn't return
// a custom value.  Tasks of this type must be wrapped with a value, either in calling
// code or via AddTaskValue.
type SimpleTask interface {
	~func() error | ~func(context.Context) error
}

// AddTaskValue wraps a SimpleTask so that it returns a value based on whether an error
// is encountered.  This function is useful when invoking functions from other packages
// that only return errors.
//
// The return task closure will invoke the SimpleTask and return success if it didn't return
// and error, and the fail value if it did.
func AddTaskValue[V any, T SimpleTask](success, fail V, raw T) func(context.Context) (V, error) {
	task, ok := any(raw).(func(context.Context) error)
	if !ok {
		f := any(raw).(func() error)
		task = func(context.Context) error {
			return f()
		}
	}

	return func(ctx context.Context) (result V, err error) {
		if err = task(ctx); err == nil {
			result = success
		} else {
			result = fail
		}

		return
	}
}
