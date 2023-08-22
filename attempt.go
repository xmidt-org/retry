package retry

import (
	"context"
	"time"
)

// Attempt represents the result of trying to invoke a task, including
// a success.
type Attempt struct {
	// Context is the policy context that spans the task attempts.
	// This field will never be nil.
	Context context.Context

	// Err is the error returned by the task.  If nil, this attempt
	// represents a success.
	Err error

	// Retries indicates the number of retries so far.  This field will
	// be zero (0) on the initial attempt.
	Retries int

	// If another retry will be attempted, this is the duration that the
	// runner will wait before the next retry.  If this field is zero (0),
	// then no further retries will be attempted.
	//
	// Use Done() to determine if this is the last attempt.  This isolates
	// client code from future changes.
	Next time.Duration
}

// Done returns true if this represents the last attempt to execute the task.
// A successful task attempt also returns true from this method, as there will
// be no more attempts.
func (a Attempt) Done() bool {
	return a.Next <= 0 || a.Context.Err() != nil
}

// OnAttempt is an optional task callback that is invoked after each attempt
// at invoking the task, including a successful one.
//
// This function must not panic or block, or task retries will be impacted.
type OnAttempt func(Attempt)
