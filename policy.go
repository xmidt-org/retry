// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package retry

import (
	"context"
	"time"
)

// PolicyFactory is a strategy for creating retry policies.  The Config
// type is an example of an implementation of this interface that can be
// unmarshaled from an external source.
type PolicyFactory interface {
	// NewPolicy creates the retry Policy this factory is configured to make.
	// This method should incorporate the context into the returned Policy.
	// Typically, this will mean a child context with a cancelation function.
	NewPolicy(context.Context) Policy
}

// PolicyFactoryFunc is a function type that implements PolicyFactory.
type PolicyFactoryFunc func(context.Context) Policy

func (pff PolicyFactoryFunc) NewPolicy(ctx context.Context) Policy { return pff(ctx) }

// Policy is a retry algorithm.  Policies are not safe for concurrent use.
// A Policy should be created each time an operation is to be executed.
type Policy interface {
	// Context returns the context associated with this policy.  This method never returns nil.
	Context() context.Context

	// Cancel halts all future retries and cleans up resources associated with this policy.
	// This method is idempotent.  After it is called the first time, Next always returns (0, false).
	Cancel()

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
	ctx        context.Context
	cancel     context.CancelFunc
	maxRetries int
	retryCount int
}

func (cp corePolicy) Context() context.Context {
	return cp.ctx
}

func (cp *corePolicy) Cancel() {
	if cp.cancel != nil {
		cp.cancel()
		cp.cancel = nil
	}
}

// withinLimits verifies that the limits of the policy, i.e. maxRetries and any context deadline,
// haven't been exceeded.  This method returns true if the policy's limits have not been
// exceeded, and false if either the limit on retries or time has been reached.
func (cp corePolicy) withinLimits() bool {
	switch {
	case cp.maxRetries > 0 && cp.retryCount >= cp.maxRetries:
		return false

	case cp.ctx.Err() != nil:
		return false

	default:
		return true
	}
}
