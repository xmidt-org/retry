package retryhttp

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/xmidt-org/retry"
)

// TaskBody is what the request body for an HTTP task must implement.  The
// io.Seeker interface is used to rewind the body between each task attempt.
//
// The request factories provided by this package ensure that a TaskBody is not
// closed by the HTTP client after each attempt.  Calling code must ensure that
// any body that is used across retries is closed if that is necessary.
//
// NOTE: bytes.Buffer does not implement this interface, but bytes.Reader does.
// Since the request body should be read-only across task attempts, bytes.Reader
// is preferred.
type TaskBody io.ReadSeeker

// taskBodyCloser wraps an io.ReadSeeker and ensures that, regardless of whether
// the ReadSeeker implements io.Closer, any call to Close is a nop.
type taskBodyCloser struct {
	TaskBody
}

func (pc taskBodyCloser) Close() error { return nil }

// RequestFactory is a closure responsible for creating a request for each task attempt.
// This closure should always incorporate the context into the request, using something
// like http.NewRequestWithContext or cloning a request with http.Request.Clone.
type RequestFactory func(context.Context) (*http.Request, error)

// NewRequest is a task-based version of http.NewRequest.  The returned factory can be
// used to create the desired request for each task attempt.
//
// If there was an error creating the request, the returned factory will return that
// error and mark it as not retryable.
func NewRequest(method, url string, body TaskBody, h http.Header) RequestFactory {
	// we take over body/getBody, so just pass nil as the body
	prototype, err := http.NewRequest(method, url, nil)
	if err != nil {
		return func(context.Context) (*http.Request, error) {
			return nil, retry.SetRetryable(err, false)
		}
	}

	// avoid clobbering the empty header set by http.NewRequest
	if len(h) > 0 {
		prototype.Header = h
	}

	return PrototypeReader(prototype, body)
}

// Prototype clones a given request for each task attempt.  This function does not
// deal with the request body.  It is only appropriate for cases where an HTTP request
// that is to be retried does not contain a body, e.g. a GET.
func Prototype(prototype *http.Request) RequestFactory {
	return func(ctx context.Context) (*http.Request, error) {
		return prototype.Clone(ctx), nil
	}
}

// PrototypeBytes clones a given request and produces a distinct Body and
// GetBody for each task attempt.
func PrototypeBytes(prototype *http.Request, body []byte) RequestFactory {
	return PrototypeReader(prototype, bytes.NewReader(body))
}

// PrototypeReader clones a given request and produces a distinct Body and
// GetBody for each task attempt.  The body may be nil.
func PrototypeReader(prototype *http.Request, b TaskBody) RequestFactory {
	var (
		contentLength int64
		body          io.ReadSeekCloser
		getBody       func() (io.ReadCloser, error)
	)

	if b != nil {
		type lengther interface {
			Len() int
		}

		if l, ok := b.(lengther); ok {
			contentLength = int64(l.Len())
		}

		body = taskBodyCloser{
			TaskBody: b,
		}

		getBody = func() (next io.ReadCloser, err error) {
			_, err = body.Seek(0, io.SeekStart)
			if err == nil {
				next = body
			}

			return
		}
	}

	return func(ctx context.Context) (request *http.Request, err error) {
		if body != nil {
			_, err = body.Seek(0, io.SeekStart)
		}

		if err == nil {
			request = prototype.Clone(ctx)
			request.ContentLength = contentLength
			request.Body = body
			request.GetBody = getBody
		}

		return
	}
}
