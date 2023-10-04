// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package retry

import "errors"

// ShouldRetryable is an interface that errors may implement to signal
// to an task runner whether an error should prevent or allow
// future retries.
type ShouldRetryable interface {
	// ShouldRetry indicates whether this error should stop all future
	// retries.
	ShouldRetry() bool
}

type shouldRetryableWrapper struct {
	error
	retryable bool
}

func (srw shouldRetryableWrapper) Unwrap() error     { return srw.error }
func (srw shouldRetryableWrapper) ShouldRetry() bool { return srw.retryable }

// SetRetryable associates retryability with a given error.  The returned
// error implements ShouldRetryable, returning the given value, and provides
// an Unwrap method for the original error.
func SetRetryable(err error, retryable bool) error {
	return shouldRetryableWrapper{
		error:     err,
		retryable: retryable,
	}
}

// ShouldRetry is a predicate for determining whether a task's results
// warrant a retry.
type ShouldRetry[V any] func(V, error) bool

// CheckRetry encapsulates the logic around determining whether a task
// retry is warranted.
//
// If err == nil, this function returns false.  No error is assumed to be
// a success, so no further retries are necessary.
//
// If the ShouldRetry predicate is not nil, the result of calling that
// predicate is returned by this function.
//
// If err != nil and it implements ShouldRetryable, the result of that
// error's ShouldRetry method is returned by this function.
//
// Failing all the above tests, this function returns true.  In other words,
// in the absence of any other indications, any non-nil error indicates
// that the task should be retried.
func CheckRetry[V any](result V, err error, p ShouldRetry[V]) bool {
	var sr ShouldRetryable
	switch {
	case err == nil:
		return false // successful task completion

	case p != nil:
		return p(result, err)

	case errors.As(err, &sr):
		return sr.ShouldRetry()

	default:
		return true // assume this task should be retried
	}
}
