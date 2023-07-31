package retry

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type ConstantSuite struct {
	CommonSuite
}

func (suite *ConstantSuite) testNextMaxRetriesExceeded() {
	p := suite.requirePolicy(
		Config{
			Interval:   5 * time.Second,
			MaxRetries: 2,
		}.NewPolicy(),
	)

	suite.Equal(5*time.Second, suite.assertContinue(p.Next()))
	suite.Equal(5*time.Second, suite.assertContinue(p.Next()))
	suite.assertStopped(p.Next())
}

func (suite *ConstantSuite) testNextMaxElapsedTimeExceeded() {
	c := suite.requireConstant(
		suite.requirePolicy(
			Config{
				Interval:       5 * time.Second,
				MaxElapsedTime: 15 * time.Second,
			}.NewPolicy(),
		),
	)

	c.now = func() time.Time {
		return c.start.Add(15 * time.Second)
	}

	suite.assertStopped(c.Next())
}

func (suite *ConstantSuite) testNextNoLimit() {
	p := suite.requirePolicy(
		Config{
			Interval: 5 * time.Second,
		}.NewPolicy(),
	)

	for i := 0; i < 10; i++ {
		suite.Equal(
			5*time.Second,
			suite.assertContinue(p.Next()),
		)
	}
}

func (suite *ConstantSuite) TestNext() {
	suite.Run("MaxRetriesExceeded", suite.testNextMaxRetriesExceeded)
	suite.Run("MaxElapsedTimeExceeded", suite.testNextMaxElapsedTimeExceeded)
	suite.Run("NoLimit", suite.testNextNoLimit)
}

func TestConstant(t *testing.T) {
	suite.Run(t, new(ConstantSuite))
}
