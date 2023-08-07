package retryhttp

import (
	"context"
	"io"
	"net/http"
)

// Client is a function that can execute an HTTP transaction.
// http.Client.Do and http.RoundTripper.RoundTrip work for this.
type Client func(*http.Request) (*http.Response, error)

// Converter converts an *http.Response into an arbitrary value. This closure is
// invoked for all non-nil responses, including non-2xx responses.  That allows
// the implementation to convert non-2xx responses into errors.
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
	// Factory is the required RequestFactory that creates *http.Request objects for
	// each execution.
	//
	// If this is nil, the task will panic.
	Factory RequestFactory

	// Client is the closure that executes HTTP transactions.  This field is optional.
	//
	// If nil, http.DefaultClient.Do is used.
	Client Client

	// Converter is the optional strategy for converting *http.Response instances into
	// the target Value type.
	//
	// If this field is nil, then the zero value of V is returned along with a nil
	// error for a successful task execution.
	Converter Converter[V]
}

// client returns Client if set, http.Default.Do otherwise.
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
//   - if the Client returned a response and there is a Converter, the Converter is invoked and its values are returned
//   - if the Client returned a response but there is no Converter, the zero value of V is returned with a nil error
//
// This method is intended for use with a retry.RunnerWithData:
//
//	var r RunnerWithData[bool] = ...
//	r.RunCtx(
//	    Task[bool]{Converter: BoolConverter}.DoCtx,
//	)
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
