package retryhttp

import (
	"context"
	"io"
	"net/http"
)

// Client is a function that can execute an HTTP transaction.
// http.Client.Do and http.RoundTripper.RoundTrip work for this.
type Client func(*http.Request) (*http.Response, error)

// RequestFactory is a closure responsible for creating a request for each task attempt.
// This closure should always incorporate the context into the request, using something
// like http.NewRequestWithContext or cloning a request with http.Request.Clone.
type RequestFactory func(context.Context) (*http.Request, error)

// Converter converts an *http.Response into an arbitrary value for further
// processing by a Task.  This closure is invoked for all non-nil responses,
// including non-2xx responses.  That allows the implementation to convert
// non-2xx responses into errors.
//
// Note: the original request is available via http.Response.Request.
type Converter[V any] func(context.Context, *http.Response) (V, error)

// cleanup handles draining and closing a client HTTP response.
func cleanup(response *http.Response) {
	if response != nil && response.Body != nil {
		io.Copy(io.Discard, response.Body)
		response.Body.Close()
	}
}

// Task represents an HTTP client task.
type Task[V any] struct {
	// Client is the closure that executes HTTP transactions.  This field is optional.
	//
	// If not supplied, http.DefaultClient.Do is used.
	Client Client

	// Factory is the required RequestFactory that creates *http.Request objects for
	// each execution.
	Factory RequestFactory

	// Converter is the optional strategy for converting *http.Response instances into
	// the target Value type.  If this field is nil, then the zero value of V is returned
	// along with a nil error for a successful task execution.
	Converter Converter[V]
}

func (t Task[V]) client() Client {
	if t.Client != nil {
		return t.Client
	}

	return http.DefaultClient.Do
}

// transact provides the backbone of HTTP tasks.  It handles creating the request
// and submitting it to a client.
func (t Task[V]) transact(ctx context.Context) (response *http.Response, err error) {
	var request *http.Request
	request, err = t.Factory(ctx)
	if err == nil {
		response, err = t.client()(request)
	}

	return
}

// DoCtx executes this HTTP task.  The basic workflow is:
//
//   - the Factory is used to create an *http.Request
//   - the Client is used to execute the request
//   - if the Client returned an error, that is returned along with the zero value of V
//   - if the Client returned a response, that response is passed to the Converter (if one is set)
func (t Task[V]) DoCtx(ctx context.Context) (result V, err error) {
	var response *http.Response
	response, err = t.transact(ctx)
	if err != nil {
		return
	}

	defer cleanup(response)
	if t.Converter != nil {
		result, err = t.Converter(ctx, response)
	}

	return
}
