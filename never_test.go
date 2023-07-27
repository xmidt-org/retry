package retry

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// NeverSuite holds unit tests for the behavior of the never policy.
type NeverSuite struct {
	CommonSuite
}

func (suite *NeverSuite) TestNext() {
	p := suite.requirePolicy(
		Config{}.NewPolicy(),
	)

	suite.assertStopped(p.Next())
}

func TestNever(t *testing.T) {
	suite.Run(t, new(NeverSuite))
}
