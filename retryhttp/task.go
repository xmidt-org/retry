package retryhttp

import (
	"context"
	"net/http"

	"github.com/xmidt-org/retry"
)

type task struct {
	client  Client
	factory RequestFactory
}

// transact provides the backbone of HTTP tasks.  It handles creating the request
// and submitting it to a client.
func (t *task) transact(ctx context.Context) (response *http.Response, err error) {
	var request *http.Request
	request, err = t.factory(ctx)
	if err == nil {
		response, err = t.client(request)
	}

	return
}

type TaskOption func(*task)

func WithClient(c Client) TaskOption {
	return func(t *task) {
		t.client = c
	}
}

func WithRequestFactory(f RequestFactory) TaskOption {
	return func(t *task) {
		t.factory = f
	}
}

func newTask(opts ...TaskOption) *task {
	t := &task{
		client: http.DefaultClient.Do,
	}

	for _, o := range opts {
		o(t)
	}

	return t
}

// NewSimpleTask creates a closure that repeatedly executes a given HTTP transaction.
// At a minimum, WithRequestFactory must appear in the options.
//
// The returned closure must be wrapped to return an arbitrary result of type V so
// that it can be passed to a Runner[V].  The retry package has several utility
// functions for doing this.
func NewSimpleTask(opts ...TaskOption) func(context.Context) error {
	t := newTask(opts...)
	return func(ctx context.Context) error {
		response, err := t.transact(ctx)
		cleanup(response)
		return err
	}
}

// NewTask creates a closure that repeatedly executes a given HTTP transaction,
// using the supplied converter strategy to convert the response into an arbitrary
// instance of type V.  At a minimum, WithRequestFactory must appear in the options.
//
// The returned task may be passed directly to a retry.Runner[V].
func NewTask[V any](c Converter[V], opts ...TaskOption) retry.Task[V] {
	t := newTask(opts...)
	return func(ctx context.Context) (result V, err error) {
		var response *http.Response
		response, err = t.transact(ctx)
		if err == nil {
			return
		}

		defer cleanup(response)
		result, err = c(ctx, response)
		return
	}
}
