package retryhttp

import (
	"bytes"
	"context"
	"io"
	"net/http"
)

// Prototype clones a given request for each task attempt.  Note that
// cloning does not handle the body.
func Prototype(prototype *http.Request) RequestFactory {
	return func(ctx context.Context) (*http.Request, error) {
		return prototype.Clone(ctx), nil
	}
}

// PrototypeBytes clones a given request and produces a distinct Body and
// GetBody for each task attempt.
func PrototypeBytes(prototype *http.Request, body []byte) RequestFactory {
	return func(ctx context.Context) (*http.Request, error) {
		request := prototype.Clone(ctx)
		request.Body = io.NopCloser(
			bytes.NewReader(body),
		)

		request.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(
				bytes.NewReader(body),
			), nil
		}

		return request, nil
	}
}

// PrototypeReader clones a given request and produces a distinct Body and
// GetBody for each task attempt.
//
// This function is a bit more efficient than PrototypeBytes, as it will reuse
// the same body by seeking to the beginning before each task attempt.  The associated
// GetBody function will also reuse the body by seeking in the same manner.
func PrototypeReader(prototype *http.Request, body io.ReadSeeker) RequestFactory {
	return func(ctx context.Context) (*http.Request, error) {
		request := prototype.Clone(ctx)

		body.Seek(0, io.SeekStart)
		request.Body = io.NopCloser(body)
		request.GetBody = func() (io.ReadCloser, error) {
			body.Seek(0, io.SeekStart)
			return io.NopCloser(body), nil
		}

		return request, nil
	}
}
