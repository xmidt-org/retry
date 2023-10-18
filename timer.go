package retry

import "time"

// Timer is a closure strategy for starting a timer.  The returned stop
// function has the same semantics as time.Timer.Stop.
//
// The default Timer used internally delegates to time.NewTimer.  A custom
// Timer is primarily useful in unit tests.
type Timer func(time.Duration) (ch <-chan time.Time, stop func() bool)

// defaultTimer is the strategy used to create a timer using the stdlib.
func defaultTimer(d time.Duration) (<-chan time.Time, func() bool) {
	t := time.NewTimer(d)
	return t.C, t.Stop
}

func nopStop() bool { return true }

// immediateTimer is a Timer function that returns a channel that is immediately
// signaled along with a stop function that is a nop.
//
// The main use case for this type of Timer is a unit test.
func immediateTimer(d time.Duration) (<-chan time.Time, func() bool) {
	ch := make(chan time.Time, 1)
	ch <- time.Now().Add(d)
	return ch, nopStop
}
