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

// DefaultTestErrorForRetry is the default strategy for determining whether a
// retry should occur.  This function does not consider the value result from
// a task.
//
// This function applies the following logic:
//
// - if err == nil, the task is assumed to be a success and this function returns false (no more retries)
// - if err implements ShouldRetryable, then err.ShouldRetry() is returned
// - if err supplies a Temporary() bool method, then err.Temporary() is returned
// - failing other logic, this function returns true
func DefaultTestErrorForRetry(err error) bool {
	if err == nil {
		return false // successful task result, so no retries are necessary
	}

	var sr ShouldRetryable
	if errors.As(err, &sr) {
		return sr.ShouldRetry()
	}

	type temporary interface {
		Temporary() bool
	}

	var t temporary
	if errors.As(err, &t) {
		return t.Temporary()
	}

	return true
}
