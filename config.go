// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package retry

import (
	"context"
	"math/rand"
	"time"
)

// Config represents the possible options when creating a Policy.  This type is friendly
// to being unmarshaled from external sources.
//
// Three basic kinds of retry policies are created by this type:
//
//   - if Interval is nonpositive, the created policy will never retry anything
//   - if Interval is positive but Jitter and Multiplier are not, the created policy will return a constant, unchanging retry interval
//   - if Interval is positive and Jitter or Multiplier are as well, the created policy will return an exponentially increasing retry interval
type Config struct {
	// Interval specifies the retry interval for a constant backoff and the
	// initial, starting interval for an exponential backoff.
	//
	// If this field is unset, no retries will happen.
	Interval time.Duration `json:"interval" yaml:"interval"`

	// Jitter is the random jitter for an exponential backoff.
	//
	// If this value is nonpositive, it is ignored.  If both this field and
	// Multiplier are unset, the resulting policy will be a constant backoff.
	Jitter float64 `json:"jitter" yaml:"jitter"`

	// Multiplier is the interval multiplier for an exponential backoff.
	//
	// If this value is less than or equal to 1.0, it is ignored.  If both this field and
	// Jitter are unset, the resulting policy will be a constant backoff.
	Multiplier float64 `json:"multiplier" yaml:"multiplier"`

	// MaxRetries is the absolute maximum number of retries performed, regardless
	// of other fields.  If this field is nonpositive, operations are retried until
	// they succeed.
	MaxRetries int `json:"maxRetries" yaml:"maxRetries"`

	// MaxElapsedTime is the absolute amount of time an operation and its retries are
	// allowed to take before giving up.  If this field is nonpositive, no maximum
	// elapsed time is enforced.
	MaxElapsedTime time.Duration `json:"maxElapsedtime" yaml:"maxElapsedTime"`

	// MaxInterval is the upper limit for each retry interval for an exponential backoff.
	// If Jitter and Multiplier are unset, or if this value is smaller than Interval, then
	// this field is ignored.
	MaxInterval time.Duration `json:"maxInterval" yaml:"maxInterval"`
}

func (c Config) newPolicyCtx(parentCtx context.Context) (context.Context, context.CancelFunc) {
	if c.MaxElapsedTime > 0 {
		return context.WithTimeout(parentCtx, c.MaxElapsedTime)
	}

	return context.WithCancel(parentCtx)
}

// NewPolicy implements PolicyFactory and uses this configuration to create the type
// of retry policy indicated by the Interval, Jitter, and Multiplier fields.
func (c Config) NewPolicy(parentCtx context.Context) Policy {
	ctx, cancel := c.newPolicyCtx(parentCtx)
	if c.Interval <= 0 {
		return &never{
			ctx:    ctx,
			cancel: cancel,
		}
	}

	cp := corePolicy{
		ctx:        ctx,
		cancel:     cancel,
		maxRetries: c.MaxRetries,
	}

	if c.Jitter <= 0.0 && c.Multiplier <= 1.0 {
		// constant is a slightly more efficient policy.
		// if the caller doesn't want randomness or an increasing interval,
		// don't make her pay the performance costs.
		return &constant{
			corePolicy: cp,
			interval:   c.Interval,
		}
	}

	return &exponential{
		corePolicy:  cp,
		rand:        rand.Int63n,
		initial:     c.Interval,
		jitter:      c.Jitter,
		multiplier:  c.Multiplier,
		maxInterval: c.MaxInterval,
	}
}
