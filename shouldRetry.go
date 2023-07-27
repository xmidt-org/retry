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

// ShouldRetry encapsulates the logic around whether an error should
// prevent further retries.  Given an err and a predicate, the following
// logic is applied, in order:
//
//   - if err == nil, this function returns false (i.e. the operation succeeded, so retries make no sense)
//   - if err or anything in its error chain implements ShouldRetryable, then err.ShouldRetry() is returned
//   - if p != nil, then p(err) is returned
//   - for a non-nil error with none of the above ways to determine retryability, this function returns false
func ShouldRetry(err error, p func(error) bool) bool {
	var sr ShouldRetryable

	switch {
	case err == nil:
		return false

	case errors.As(err, &sr):
		return sr.ShouldRetry()

	case p != nil:
		return p(err)

	default:
		return false
	}
}
