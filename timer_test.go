package retry

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type TimerSuite struct {
	suite.Suite
}

// TestDefaultTimer is just a smoke test to make sure the defaultTimer
// operates basically as intended.
func (suite *TimerSuite) TestDefaultTimer() {
	ch, stop := defaultTimer(100 * time.Millisecond)
	suite.NotNil(ch)
	suite.Require().NotNil(stop)

	stop()
	stop() // idempotent
}

func (suite *TimerSuite) TestImmediateTimer() {
	ch, stop := immediateTimer(5 * time.Second)
	suite.Require().NotNil(ch)
	suite.Require().NotNil(stop)

	select {
	case <-ch:
		// passing
	case <-time.After(time.Second):
		suite.Fail("immediate timer channel was not signaled")
	}

	stop()
	stop() // idempotent
}

func TestTimer(t *testing.T) {
	suite.Run(t, new(TimerSuite))
}
