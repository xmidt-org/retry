package retry

import (
	"time"
)

// exponential is the main implementing type for Policy.
type exponential struct {
	corePolicy
	rand        func(int64) int64
	initial     time.Duration
	previous    time.Duration
	jitter      float64
	multiplier  float64
	maxInterval time.Duration
}

// nextBaseInterval computes the next un-jittered retry interval
// and sets up the next interval to use.  On the first call, this
// method simply returns the initial interval.  Subsequent calls
// return intervals that grow exponentially using the multiplier as
// a base.  If no multiplier is set, this method just returns the
// initial interval every time.
func (e *exponential) nextBaseInterval() (base time.Duration) {
	if e.previous > 0 {
		base = e.previous

		if e.multiplier > 1.0 {
			base = time.Duration(float64(base) * e.multiplier)
		}

		if e.maxInterval > 0 && base > e.maxInterval {
			base = e.maxInterval
		}
	} else {
		base = e.initial
	}

	e.previous = base
	return
}

// jitterize computes a random interval using the jitter value.  If jitter is
// nonpositive, this method returns base as is.
func (e *exponential) jitterize(base time.Duration) (next time.Duration) {
	next = base
	if e.jitter > 0.0 {
		delta := int64(float64(next) * e.jitter)

		// choose a random value in the range [next-delta, next+delta]
		next = next - time.Duration(delta) + time.Duration(e.rand(2*delta+1))
	}

	if e.maxInterval > 0 && next > e.maxInterval {
		next = e.maxInterval
	}

	return
}

func (e *exponential) Next() (time.Duration, bool) {
	if !e.withinLimits() {
		return 0, false
	}

	e.retryCount++
	return e.jitterize(
		e.nextBaseInterval(),
	), true
}
