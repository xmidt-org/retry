package retryhttp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type ClientSuite struct {
	suite.Suite
}

func (suite *ClientSuite) timerStop() bool {
	return true
}

// timer is a retry.Timer that returns a time channel that is immediately signaled.
func (suite *ClientSuite) timer(d time.Duration) (<-chan time.Time, func() bool) {
	suite.True(d > 0)
	ch := make(chan time.Time, 1)
	ch <- time.Now()
	return ch, suite.timerStop
}

func (suite *ClientSuite) TestDo() {
}

func TestClient(t *testing.T) {
	suite.Run(t, new(ClientSuite))
}
