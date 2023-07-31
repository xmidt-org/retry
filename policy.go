package retry

import (
	"time"
)

// PolicyFactory is a strategy for creating retry policies.  The Config
// type is an example of an implementation of this interface that can be
// unmarshaled from an external source.
type PolicyFactory interface {
	// NewPolicy creates the retry Policy this factory is configured to make.
	NewPolicy() Policy
}

// PolicyFactoryFunc is a function type that implements PolicyFactory.
type PolicyFactoryFunc func() Policy

func (pff PolicyFactoryFunc) NewPolicy() Policy { return pff() }

// Policy is a retry algorithm.  Policies are not safe for concurrent use.
// A Policy should be created each time an operation is to be executed.
type Policy interface {
	// Next obtains the retry interval to use next.  Typically, a caller will
	// sleep for the returned duration before trying the operation again.
	//
	// This method is not safe for concurrent invocation.
	//
	// If this method returns true, the duration will be a positive value indicating
	// the amount of time that should elapse before the next retry.
	//
	// If this method returns false, the duration will always be zero (0).
	Next() (time.Duration, bool)
}

// corePolicy implements the common functionality for policies other than those
// that never retry.
type corePolicy struct {
	maxRetries     int
	maxElapsedTime time.Duration
	now            func() time.Time
	start          time.Time
	retryCount     int
}

// withinLimits verifies that the limits of the policy, i.e. maxRetries and maxElapsedTime,
// haven't been exceeded.  This method returns true if the policy's limits have not been
// exceeded, and false if either the limit on retries or time has been reached.
func (cp corePolicy) withinLimits() bool {
	switch {
	case cp.maxRetries > 0 && cp.retryCount >= cp.maxRetries:
		return false

	case cp.maxElapsedTime > 0 && cp.now().Sub(cp.start) >= cp.maxElapsedTime:
		return false

	default:
		return true
	}
}
